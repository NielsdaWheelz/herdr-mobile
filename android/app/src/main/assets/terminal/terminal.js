(function () {
    "use strict";

    var terminalHost = document.getElementById("terminal");
    var terminalStatus = document.getElementById("terminal-status");
    var pagePort = null;
    var pageFailed = false;
    var fitScheduled = false;
    var fontsSettled = false;
    var modifiers = { control: "Off", alt: "Off" };
    var compositionActive = false;
    var maximumInputBytes = 32 * 1024;
    var maximumSelectionBytes = 256 * 1024;
    var maximumFrameBytes = 2 * 1024 * 1024;
    // The gateway's Resize bounds; the page fits down to them and never up.
    var minimumColumns = 20;
    var maximumColumns = 1024;
    var minimumRows = 5;
    var maximumRows = 512;
    var lastPublishedColumns = 0;
    var lastPublishedRows = 0;
    var viewportTooSmallPublished = false;
    var terminal = null;
    var terminalTouchInteraction = null;

    window.addEventListener("message", acceptPagePort);
    if (!terminalHost || !terminalStatus) {
        failPage();
        return;
    }

    // xterm reads extendedAnsi as one array anchored at ansi index 16, so the
    // 24-step grayscale (indices 232-255) lands at offsets 216-239 and the
    // 6x6x6 cube keeps its library defaults.
    var ink = [0x0c, 0x0d, 0x0f];
    var bone = [0xf3, 0xf0, 0xe8];
    var extendedAnsi = new Array(240);
    for (var step = 0; step < 24; step += 1) {
        extendedAnsi[216 + step] = "#" + ink.map(function (channel, index) {
            var tone = Math.round(channel + (bone[index] - channel) * step / 23);
            return (tone < 16 ? "0" : "") + tone.toString(16);
        }).join("");
    }

    terminal = new window.Terminal({
        cursorBlink: true,
        fontFamily: '"JetBrains Mono", monospace',
        minimumContrastRatio: 3,
        rows: 8,
        scrollback: 1000,
        screenReaderMode: true,
        theme: {
            background: "#0c0d0f",
            foreground: "#f3f0e8",
            cursor: "#d6a85f",
            cursorAccent: "#0c0d0f",
            selectionBackground: "#f3f0e84d",
            selectionInactiveBackground: "#f3f0e826",
            overviewRulerBorder: "#aaa69d",
            black: "#15171a",
            red: "#d74e33",
            green: "#4f925c",
            yellow: "#ac7e35",
            blue: "#538bac",
            magenta: "#bb5897",
            cyan: "#459c93",
            white: "#aaa69d",
            brightBlack: "#5c6370",
            brightRed: "#e46c55",
            brightGreen: "#76b082",
            brightYellow: "#d6a85f",
            brightBlue: "#78a9c6",
            brightMagenta: "#cd70ab",
            brightCyan: "#64c4ba",
            brightWhite: "#f3f0e8",
            extendedAnsi: extendedAnsi
        }
    });
    var fitAddon = new window.FitAddon.FitAddon();
    terminal.loadAddon(fitAddon);
    terminal.open(terminalHost);

    function send(payload) {
        if (pagePort) pagePort.postMessage(JSON.stringify(payload));
    }

    function failPage() {
        if (pageFailed) return;
        pageFailed = true;
        if (terminalTouchInteraction) terminalTouchInteraction.dispose();
        setModifiers("Off", "Off");
        send({ kind: "PageFailure" });
        pagePort = null;
    }

    function publishModifiers() {
        send({
            kind: "ModifierState",
            control: modifiers.control,
            alt: modifiers.alt
        });
    }

    function setModifiers(control, alt) {
        if (modifiers.control === control && modifiers.alt === alt) return;
        modifiers = { control: control, alt: alt };
        publishModifiers();
    }

    function resetInputState() {
        if (terminalTouchInteraction) terminalTouchInteraction.cancel();
        setModifiers("Off", "Off");
    }

    function utf8ByteCount(value, maximumBytes) {
        if (value.length > maximumBytes) return null;
        var count = 0;
        for (var index = 0; index < value.length; index += 1) {
            var first = value.charCodeAt(index);
            var code = first;
            if (first >= 0xd800 && first <= 0xdbff) {
                var second = value.charCodeAt(index + 1);
                if (second < 0xdc00 || second > 0xdfff) return null;
                code = value.codePointAt(index);
                index += 1;
            } else if (first >= 0xdc00 && first <= 0xdfff) {
                return null;
            }
            count += code <= 0x7f ? 1 : code <= 0x7ff ? 2 : code <= 0xffff ? 3 : 4;
            if (count > maximumBytes) return null;
        }
        return count;
    }

    function isUnicodeScalarSequence(value) {
        for (var index = 0; index < value.length; index += 1) {
            var first = value.charCodeAt(index);
            if (first >= 0xd800 && first <= 0xdbff) {
                var second = value.charCodeAt(index + 1);
                if (second < 0xdc00 || second > 0xdfff) return false;
                index += 1;
            } else if (first >= 0xdc00 && first <= 0xdfff) {
                return false;
            }
        }
        return true;
    }

    function sendText(value) {
        if (typeof value !== "string" || value.length === 0 ||
            /[\u0000-\u001f\u007f]/.test(value) ||
            utf8ByteCount(value, maximumInputBytes) === null) {
            terminalStatus.textContent = "Text unavailable with this terminal";
            return;
        }
        setModifiers("Off", "Off");
        send({ kind: "Text", text: value });
    }

    function sendKey(key, shift, physicalControl, physicalAlt) {
        var armed = modifiers;
        setModifiers("Off", "Off");
        var keyModifiers = [];
        if (armed.control === "Armed" || physicalControl) keyModifiers.push("ctrl");
        if (armed.alt === "Armed" || physicalAlt) keyModifiers.push("alt");
        if (shift) keyModifiers.push("shift");
        send({ kind: "Key", key: key, modifiers: keyModifiers });
    }

    function pasteInput(value) {
        if (value.length === 0) return;
        setModifiers("Off", "Off");
        send({ kind: "Paste", text: value });
    }

    function sanitizePaste(value) {
        value = String(value);
        if (value.length > maximumInputBytes) return null;
        var sanitized = "";
        var byteCount = 0;
        for (var index = 0; index < value.length; index += 1) {
            var code = value.codePointAt(index);
            var character = String.fromCodePoint(code);
            if (code > 0xffff) index += 1;
            if (code === 0x0d) {
                if (value.charCodeAt(index + 1) === 0x0a) index += 1;
                code = 0x0a;
                character = "\n";
            }
            if (code >= 0xd800 && code <= 0xdfff) continue;
            if (code === 0x09 || code === 0x0a || (code > 0x1f && code < 0x7f) || code > 0x9f) {
                byteCount += code <= 0x7f ? 1 : code <= 0x7ff ? 2 : code <= 0xffff ? 3 : 4;
                if (byteCount > maximumInputBytes) return null;
                sanitized += character;
            }
        }
        return sanitized;
    }

    // Whole cells at the native-chosen font, capped downward at the gateway
    // maximum and never clamped upward. Below the gateway minimum xterm keeps
    // its last valid grid and native hears ViewportTooSmall once per episode;
    // recovery republishes Resize even when the grid is unchanged.
    function resizeTerminal() {
        fitScheduled = false;
        var dimensions = terminalHost.clientWidth > 0 && terminalHost.clientHeight > 0 ?
            fitAddon.proposeDimensions() : undefined;
        if (!dimensions || dimensions.cols < minimumColumns || dimensions.rows < minimumRows) {
            if (fontsSettled && pagePort && !viewportTooSmallPublished) {
                viewportTooSmallPublished = true;
                send({ kind: "ViewportTooSmall" });
            }
            return;
        }
        var columns = Math.min(maximumColumns, dimensions.cols);
        var rows = Math.min(maximumRows, dimensions.rows);
        if (terminal.cols !== columns || terminal.rows !== rows) {
            if (terminalTouchInteraction) terminalTouchInteraction.cancel();
            terminal.resize(columns, rows);
        }
        if (!fontsSettled || !pagePort) return;
        if (viewportTooSmallPublished || columns !== lastPublishedColumns || rows !== lastPublishedRows) {
            viewportTooSmallPublished = false;
            lastPublishedColumns = columns;
            lastPublishedRows = rows;
            send({ kind: "Resize", columns: columns, rows: rows });
        }
    }

    function scheduleFit() {
        if (fitScheduled) return;
        fitScheduled = true;
        window.requestAnimationFrame(resizeTerminal);
    }

    function isFontSizeCssPx(value) {
        return typeof value === "number" && Number.isFinite(value) && value > 0;
    }

    function exactObject(value, expectedKeys) {
        if (!value || typeof value !== "object" || Array.isArray(value)) return false;
        var actualKeys = Object.keys(value);
        if (actualKeys.length !== expectedKeys.length) return false;
        return expectedKeys.every(function (key) {
            return Object.prototype.hasOwnProperty.call(value, key);
        });
    }

    function parseObject(value) {
        if (typeof value !== "string") return null;
        try {
            var parsed = JSON.parse(value);
            return parsed && typeof parsed === "object" && !Array.isArray(parsed) ? parsed : null;
        } catch (error) {
            return null;
        }
    }

    function decodeBase64(value) {
        if (typeof value !== "string" || value.length > Math.ceil(maximumFrameBytes / 3) * 4 ||
            value.length % 4 !== 0 ||
            !/^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$/.test(value)) {
            return null;
        }
        try {
            var decoded = window.atob(value);
            if (decoded.length > maximumFrameBytes) return null;
            var bytes = new Uint8Array(decoded.length);
            for (var index = 0; index < decoded.length; index += 1) bytes[index] = decoded.charCodeAt(index);
            return bytes;
        } catch (error) {
            return null;
        }
    }

    function rejectKey(event) {
        event.preventDefault();
        event.stopImmediatePropagation();
        terminalStatus.textContent = "Key unavailable with this terminal";
    }

    function acceptHardwareKey(event) {
        if (pageFailed || event.target !== input || event.isTrusted !== true ||
            compositionActive || event.isComposing || event.keyCode === 229) return;
        var name = event.key;
        if (typeof name !== "string" || name === "Process" || name === "Unidentified") return;
        if (name === "Home" || name === "End" || name === "Insert" || name === "Delete") {
            rejectKey(event);
            return;
        }
        if (name === "PageUp" || name === "PageDown") {
            if (event.ctrlKey || event.altKey || event.shiftKey || event.metaKey ||
                modifiers.control === "Armed" || modifiers.alt === "Armed") {
                rejectKey(event);
                return;
            }
            event.preventDefault();
            event.stopImmediatePropagation();
            send({
                kind: "Scroll", source: "PageKey",
                direction: name === "PageUp" ? "Up" : "Down", lines: terminal.rows
            });
            return;
        }
        var names = {
            Enter: "enter", Escape: "escape", Tab: "tab", Backspace: "backspace",
            ArrowUp: "up", ArrowDown: "down", ArrowLeft: "left", ArrowRight: "right"
        };
        var key = names[name];
        if (key === undefined && /^F(?:[1-9]|1[0-2])$/.test(name)) key = name.toLowerCase();
        var modified = event.ctrlKey || event.altKey || modifiers.control === "Armed" ||
            modifiers.alt === "Armed";
        if (key === undefined && Array.from(name).length === 1 && modified) {
            if (name !== " " && (/^[\p{White_Space}\p{Zs}]$/u.test(name) ||
                /^[\u0000-\u001f\u007f-\u009f]$/.test(name))) {
                rejectKey(event);
                return;
            }
            key = name;
        }
        if (key === undefined) {
            if (name === "Dead" || name === "Compose" ||
                name === "Shift" || name === "Control" || name === "Alt" ||
                name === "Meta" || name === "AltGraph" ||
                name === "CapsLock" || name === "NumLock") return;
            if (event.metaKey || Array.from(name).length !== 1) rejectKey(event);
            return;
        }
        if (event.metaKey) {
            rejectKey(event);
            return;
        }
        event.preventDefault();
        event.stopImmediatePropagation();
        sendKey(key, event.shiftKey, event.ctrlKey, event.altKey);
    }

    function createTerminalTouchInteraction(options) {
        var ownerTerminal = options.terminal;
        var ownerScreen = options.screen;
        var longPressMilliseconds = options.longPressMilliseconds;
        var ownerInput = document.querySelector(".xterm-helper-textarea");
        var touchSlop = 8;
        var gesture = null;
        var selectionGeneration = null;
        var lastAcknowledgedSelectionGeneration = null;
        var nextSelectionGeneration = 1;
        var disposed = false;

        function canonicalSelectionGeneration(value) {
            if (typeof value !== "string" || !/^[1-9][0-9]*$/.test(value)) return null;
            var numeric = Number(value);
            return Number.isSafeInteger(numeric) && numeric > 0 && String(numeric) === value ?
                value : null;
        }

        function compareSelectionGenerations(left, right) {
            if (left.length !== right.length) return left.length < right.length ? -1 : 1;
            return left === right ? 0 : left < right ? -1 : 1;
        }

        function beginSelectionGeneration() {
            if (selectionGeneration !== null && !selectionGeneration.clearAcknowledged) {
                failPage();
                return null;
            }
            if (!Number.isSafeInteger(nextSelectionGeneration) || nextSelectionGeneration <= 0) {
                failPage();
                return null;
            }
            var generation = String(nextSelectionGeneration);
            nextSelectionGeneration += 1;
            selectionGeneration = { value: generation, clearAcknowledged: false };
            return generation;
        }

        function isTrustedTouch(event) {
            return event.isTrusted === true;
        }

        function isScreenTouch(event) {
            var target = event.target;
            return target === ownerScreen ||
                (target && target.nodeType === Node.ELEMENT_NODE && ownerScreen.contains(target));
        }

        function containIgnoredTouch(event) {
            if (event.cancelable) event.preventDefault();
            event.stopImmediatePropagation();
        }

        function consumeTouch(event, requireCancelable) {
            if (requireCancelable && !event.cancelable) {
                event.stopImmediatePropagation();
                failPage();
                return false;
            }
            if (event.cancelable) {
                event.preventDefault();
                if (!event.defaultPrevented) {
                    event.stopImmediatePropagation();
                    failPage();
                    return false;
                }
            }
            event.stopImmediatePropagation();
            return true;
        }

        function clearFrame(active) {
            if (active.frame === null || active.frame === undefined) return;
            window.cancelAnimationFrame(active.frame);
            active.frame = null;
        }

        function clearTimer(active) {
            if (active.timer === null || active.timer === undefined) return;
            window.clearTimeout(active.timer);
            active.timer = null;
        }

        function findTouch(list, identifier) {
            var found = null;
            for (var index = 0; index < list.length; index += 1) {
                if (list[index].identifier !== identifier) continue;
                if (found !== null) return null;
                found = list[index];
            }
            return found;
        }

        function acknowledgeSelectionCleared() {
            if (selectionGeneration === null) {
                failPage();
                return false;
            }
            if (!selectionGeneration.clearAcknowledged) {
                selectionGeneration.clearAcknowledged = true;
                lastAcknowledgedSelectionGeneration = selectionGeneration.value;
                send({
                    kind: "SelectionCleared",
                    generation: selectionGeneration.value
                });
            }
            return true;
        }

        function cancelGesture(blockTail) {
            var active = gesture;
            if (active === null) return;
            clearTimer(active);
            clearFrame(active);
            var clearsSemanticDrag = active.state === "Selecting";
            var clearsReleasedSelection = active.state === "Selected" ||
                active.state === "PendingSelected" || active.state === "BlockingSelected";
            if (active.state === "Scrolling") {
                active.accumulator = 0;
                active.emittedRows = 0;
            }
            gesture = blockTail && active.identifier !== undefined ?
                { state: "Blocked", identifier: active.identifier, frame: null, timer: null } : null;
            try {
                if (clearsSemanticDrag) {
                    ownerTerminal.handleSelectionInput({ kind: "Cancel" });
                } else if (clearsReleasedSelection) {
                    ownerTerminal.clearSelection();
                }
            } catch (error) {
                failPage();
                return;
            }
            if ((clearsSemanticDrag || clearsReleasedSelection) && !acknowledgeSelectionCleared()) return;
        }

        function clearSelectionFromNative() {
            var generation = arguments[0];
            if (canonicalSelectionGeneration(generation) === null) {
                failPage();
                return;
            }
            if (selectionGeneration === null || generation !== selectionGeneration.value) {
                if (lastAcknowledgedSelectionGeneration !== null &&
                    compareSelectionGenerations(generation, lastAcknowledgedSelectionGeneration) <= 0) return;
                failPage();
                return;
            }
            if (selectionGeneration.clearAcknowledged) return;
            var active = gesture;
            if (active !== null) {
                cancelGesture(true);
                if (pageFailed) return;
                if (active.state === "Selecting" || active.state === "Selected" ||
                    active.state === "PendingSelected" || active.state === "BlockingSelected") return;
            }
            try {
                ownerTerminal.clearSelection();
            } catch (error) {
                failPage();
                return;
            }
            acknowledgeSelectionCleared();
        }

        function sendWheelLines(deltaLines, clientX, clientY) {
            if (deltaLines === 0) return;
            var bounds = ownerScreen.getBoundingClientRect();
            if (bounds.width <= 0 || bounds.height <= 0 ||
                ownerTerminal.cols <= 0 || ownerTerminal.rows <= 0) return;
            var column = Math.max(0, Math.min(
                ownerTerminal.cols - 1,
                Math.floor((clientX - bounds.left) * ownerTerminal.cols / bounds.width)
            ));
            var row = Math.max(0, Math.min(
                ownerTerminal.rows - 1,
                Math.floor((clientY - bounds.top) * ownerTerminal.rows / bounds.height)
            ));
            send({
                kind: "Scroll", source: "Wheel",
                direction: deltaLines < 0 ? "Up" : "Down",
                lines: Math.abs(deltaLines), column: column, row: row
            });
        }

        function wholeRows(displacement, rowHeight) {
            return displacement < 0 ?
                Math.ceil(displacement / rowHeight) :
                Math.floor(displacement / rowHeight);
        }

        function dispatchAccumulatedRows(active) {
            if (gesture !== active || active.state !== "Scrolling") return;
            if (ownerTerminal.hasSelection()) {
                cancelGesture(true);
                return;
            }
            var rows = wholeRows(active.accumulator, active.rowHeight);
            if (rows === 0) return;
            var rowLimit = ownerTerminal.rows;
            if (!Number.isFinite(rowLimit) || rowLimit <= 0) {
                cancelGesture(true);
                return;
            }
            var boundedRows = Math.max(-rowLimit, Math.min(rowLimit, rows));
            active.accumulator -= rows * active.rowHeight;
            active.emittedRows += boundedRows;
            sendWheelLines(boundedRows, active.clientX, active.clientY);
        }

        function scheduleDispatch(active) {
            if (active.frame !== null) return;
            active.frame = window.requestAnimationFrame(function () {
                active.frame = null;
                dispatchAccumulatedRows(active);
            });
        }

        function addMovement(active, currentY) {
            if (!Number.isFinite(currentY)) {
                cancelGesture(true);
                return false;
            }
            active.accumulator += active.previousY - currentY;
            active.previousY = currentY;
            return true;
        }

        function startSelection(active) {
            clearTimer(active);
            if (active.state === "PendingSelected") {
                try {
                    ownerTerminal.clearSelection();
                } catch (error) {
                    gesture = null;
                    failPage();
                    return;
                }
                if (!acknowledgeSelectionCleared()) return;
            }
            var generation = beginSelectionGeneration();
            if (generation === null) return;
            try {
                ownerTerminal.handleSelectionInput({
                    kind: "StartWord",
                    clientX: active.startX,
                    clientY: active.startY
                });
            } catch (error) {
                gesture = null;
                failPage();
                return;
            }
            active.state = "Selecting";
            send({ kind: "SelectionStarted", generation: generation });
        }

        function extendSelection(active, kind, clientX, clientY) {
            try {
                ownerTerminal.handleSelectionInput({
                    kind: kind,
                    clientX: clientX,
                    clientY: clientY
                });
                return true;
            } catch (error) {
                failPage();
                return false;
            }
        }

        function releaseSelection(active, touch) {
            if (!extendSelection(active, "End", touch.clientX, touch.clientY)) return;
            gesture = null;
            var text = ownerTerminal.getSelection();
            if (typeof text !== "string" || !isUnicodeScalarSequence(text)) {
                ownerTerminal.clearSelection();
                failPage();
                return;
            }
            if (text.length === 0) {
                ownerTerminal.clearSelection();
                acknowledgeSelectionCleared();
                return;
            }
            if (utf8ByteCount(text, maximumSelectionBytes) === null) {
                ownerTerminal.clearSelection();
                send({
                    kind: "SelectionCopyRejected",
                    generation: selectionGeneration.value,
                    reason: "TooLarge"
                });
                return;
            }
            var width = window.innerWidth;
            var height = window.innerHeight;
            if (!Number.isFinite(width) || width <= 0 || !Number.isFinite(height) || height <= 0) {
                ownerTerminal.clearSelection();
                failPage();
                return;
            }
            gesture = { state: "Selected", frame: null, timer: null };
            send({
                kind: "SelectionAvailable",
                generation: selectionGeneration.value,
                anchorX: Math.max(0, Math.min(1, touch.clientX / width)),
                anchorY: Math.max(0, Math.min(1, touch.clientY / height)),
                text: text
            });
        }

        function touchIsInsideScreen(touch) {
            var bounds = ownerScreen.getBoundingClientRect();
            return Number.isFinite(touch.clientX) && Number.isFinite(touch.clientY) &&
                touch.clientX > bounds.left && touch.clientX < bounds.right &&
                touch.clientY > bounds.top && touch.clientY < bounds.bottom;
        }

        function blockGesture(event) {
            cancelGesture(true);
            if (gesture === null) {
                gesture = { state: "Blocked", frame: null, timer: null };
            }
            if (event.touches.length === 0) gesture = null;
        }

        function onTouchStart(event) {
            if (disposed) return;
            var ownsActiveStream = gesture !== null && gesture.state !== "Selected";
            if (!ownsActiveStream && !isScreenTouch(event)) return;
            if (!isTrustedTouch(event)) {
                containIgnoredTouch(event);
                return;
            }
            if (!consumeTouch(event, !ownsActiveStream)) return;
            var retainedSelection = gesture !== null && gesture.state === "Selected";
            if (gesture !== null && !retainedSelection) {
                blockGesture(event);
                return;
            }
            if (event.touches.length !== 1 || event.changedTouches.length !== 1) {
                failPage();
                return;
            }
            var touch = event.changedTouches[0];
            touch = findTouch(event.touches, touch.identifier);
            if (touch === null || !touchIsInsideScreen(touch)) {
                failPage();
                return;
            }
            if (!retainedSelection && compositionActive) {
                gesture = {
                    state: "Composing",
                    identifier: touch.identifier,
                    frame: null,
                    timer: null
                };
                return;
            }
            var bounds = ownerScreen.getBoundingClientRect();
            var rows = ownerTerminal.rows;
            var rowHeight = bounds.height / rows;
            if (!Number.isFinite(rowHeight) || rowHeight <= 0) {
                failPage();
                return;
            }
            var horizontalInset = Math.min(0.5, bounds.width / 2);
            var verticalInset = Math.min(0.5, bounds.height / 2);
            var pending = {
                state: retainedSelection ? "PendingSelected" : "Pending",
                identifier: touch.identifier,
                startX: touch.clientX,
                startY: touch.clientY,
                previousY: touch.clientY,
                clientX: Math.max(bounds.left + horizontalInset,
                    Math.min(bounds.right - horizontalInset, touch.clientX)),
                clientY: Math.max(bounds.top + verticalInset,
                    Math.min(bounds.bottom - verticalInset, touch.clientY)),
                rowHeight: rowHeight,
                accumulator: 0,
                emittedRows: 0,
                frame: null,
                timer: null
            };
            gesture = pending;
            pending.timer = window.setTimeout(function () {
                if (gesture === pending) startSelection(pending);
            }, longPressMilliseconds);
        }

        function onTouchMove(event) {
            if (disposed) return;
            var active = gesture;
            if (active === null && !isScreenTouch(event)) return;
            if (!isTrustedTouch(event)) {
                containIgnoredTouch(event);
                return;
            }
            if (active === null) {
                containIgnoredTouch(event);
                return;
            }
            if (!consumeTouch(event)) return;
            if (active.state === "Blocked") {
                if (event.touches.length === 0) gesture = null;
                return;
            }
            if (event.touches.length !== 1) {
                blockGesture(event);
                return;
            }
            var touch = findTouch(event.touches, active.identifier);
            if (touch === null) {
                blockGesture(event);
                return;
            }
            if (active.state === "Composing") return;
            if (active.state === "Selecting") {
                extendSelection(active, "Extend", touch.clientX, touch.clientY);
                return;
            }
            if (active.state === "Scrolling" || active.state === "BlockingSelected") {
                if (active.state === "Scrolling" && addMovement(active, touch.clientY)) {
                    scheduleDispatch(active);
                }
                return;
            }
            if (active.state === "Pending" || active.state === "PendingSelected") {
                var deltaX = touch.clientX - active.startX;
                var deltaY = touch.clientY - active.startY;
                var absoluteX = Math.abs(deltaX);
                var absoluteY = Math.abs(deltaY);
                if (absoluteX <= touchSlop && absoluteY <= touchSlop) return;
                clearTimer(active);
                if (active.state === "PendingSelected") {
                    active.state = "BlockingSelected";
                    return;
                }
                if (absoluteY <= absoluteX) {
                    blockGesture(event);
                    return;
                }
                active.state = "Scrolling";
            }
            if (active.state !== "Scrolling") {
                failPage();
                return;
            }
            if (addMovement(active, touch.clientY)) scheduleDispatch(active);
        }

        function onTouchEnd(event) {
            if (disposed) return;
            var active = gesture;
            if (active === null && !isScreenTouch(event)) return;
            if (!isTrustedTouch(event)) {
                containIgnoredTouch(event);
                return;
            }
            if (active === null) {
                containIgnoredTouch(event);
                return;
            }
            if (!consumeTouch(event)) return;
            if (active.state === "Blocked") {
                if (event.touches.length === 0) gesture = null;
                return;
            }
            if (event.touches.length !== 0 || event.changedTouches.length !== 1) {
                blockGesture(event);
                return;
            }
            var touch = findTouch(event.changedTouches, active.identifier);
            if (touch === null || !Number.isFinite(touch.clientX) || !Number.isFinite(touch.clientY)) {
                blockGesture(event);
                return;
            }
            if (active.state === "Composing") {
                gesture = null;
                return;
            }
            if (active.state === "Pending") {
                clearTimer(active);
                gesture = null;
                focusTerminal();
                if (pageFailed) return;
                send({ kind: "ImeRequested" });
                return;
            }
            if (active.state === "PendingSelected") {
                clearTimer(active);
                gesture = null;
                ownerTerminal.clearSelection();
                acknowledgeSelectionCleared();
                return;
            }
            if (active.state === "Selecting") {
                releaseSelection(active, touch);
                return;
            }
            if (active.state === "BlockingSelected") {
                gesture = { state: "Selected", frame: null, timer: null };
                return;
            }
            if (active.state !== "Scrolling") {
                failPage();
                return;
            }
            clearFrame(active);
            var rowLimit = ownerTerminal.rows;
            if (!Number.isFinite(rowLimit) || rowLimit <= 0) {
                cancelGesture(true);
                return;
            }
            var targetRows = wholeRows(active.startY - touch.clientY, active.rowHeight);
            targetRows = Math.max(-rowLimit, Math.min(rowLimit, targetRows));
            var correctionRows = Math.max(
                -rowLimit,
                Math.min(rowLimit, targetRows - active.emittedRows)
            );
            if (correctionRows !== 0) {
                sendWheelLines(correctionRows, active.clientX, active.clientY);
            }
            gesture = null;
            active.accumulator = 0;
            active.emittedRows = 0;
        }

        function onTouchCancel(event) {
            if (disposed) return;
            if (gesture === null && !isScreenTouch(event)) return;
            if (!isTrustedTouch(event)) {
                containIgnoredTouch(event);
                return;
            }
            if (gesture === null) {
                containIgnoredTouch(event);
                return;
            }
            if (!consumeTouch(event)) return;
            blockGesture(event);
        }

        function onInputBlur() {
            if (gesture !== null && gesture.state !== "Selected") cancelGesture(true);
        }

        function cancelActiveGesture() {
            if (gesture !== null && gesture.state !== "Selected") cancelGesture(true);
        }

        function onVisibilityChange() {
            if (document.visibilityState === "hidden") cancelGesture(true);
        }

        function scroll(direction) {
            if (disposed) return;
            if (selectionGeneration !== null && !selectionGeneration.clearAcknowledged) {
                cancelGesture(true);
                return;
            }
            if (ownerTerminal.hasSelection()) {
                failPage();
                return;
            }
            cancelGesture(true);
            var bounds = ownerScreen.getBoundingClientRect();
            var rows = ownerTerminal.rows;
            if (!Number.isFinite(bounds.width) || !Number.isFinite(bounds.height) ||
                bounds.width <= 0 || bounds.height <= 0 ||
                !Number.isFinite(rows) || rows <= 0) return;
            var magnitude = Math.max(1, rows - 1);
            sendWheelLines(
                direction === "Backward" ? -magnitude : magnitude,
                bounds.left + bounds.width / 2,
                bounds.top + bounds.height / 2
            );
        }

        function dispose() {
            if (disposed) return;
            cancelGesture(false);
            disposed = true;
            window.removeEventListener("touchstart", onTouchStart, true);
            window.removeEventListener("touchmove", onTouchMove, true);
            window.removeEventListener("touchend", onTouchEnd, true);
            window.removeEventListener("touchcancel", onTouchCancel, true);
            if (ownerInput) ownerInput.removeEventListener("blur", onInputBlur, true);
            window.removeEventListener("blur", cancelGesture, true);
            window.removeEventListener("resize", cancelActiveGesture, true);
            window.removeEventListener("orientationchange", cancelActiveGesture, true);
            window.removeEventListener("pagehide", dispose, true);
            document.removeEventListener("visibilitychange", onVisibilityChange, true);
        }

        window.addEventListener("touchstart", onTouchStart, { capture: true, passive: false });
        window.addEventListener("touchmove", onTouchMove, { capture: true, passive: false });
        window.addEventListener("touchend", onTouchEnd, { capture: true, passive: false });
        window.addEventListener("touchcancel", onTouchCancel, { capture: true, passive: false });
        if (ownerInput) ownerInput.addEventListener("blur", onInputBlur, true);
        window.addEventListener("blur", cancelGesture, true);
        window.addEventListener("resize", cancelActiveGesture, true);
        window.addEventListener("orientationchange", cancelActiveGesture, true);
        window.addEventListener("pagehide", dispose, true);
        document.addEventListener("visibilitychange", onVisibilityChange, true);

        return {
            cancel: function () { cancelGesture(true); },
            clearSelection: clearSelectionFromNative,
            scroll: scroll,
            dispose: dispose
        };
    }

    var input = document.querySelector(".xterm-helper-textarea");
    var composition = document.querySelector(".composition-view");
    var screen = document.querySelector(".xterm-screen");

    function lockPageViewport() {
        document.documentElement.scrollLeft = 0;
        document.documentElement.scrollTop = 0;
        document.body.scrollLeft = 0;
        document.body.scrollTop = 0;
        window.scrollTo(0, 0);
    }

    function containImeGeometry() {
        if (!input || !composition || !screen) {
            failPage();
            return;
        }
        var screenBounds = screen.getBoundingClientRect();
        var compositionBounds = composition.getBoundingClientRect();
        var inputBounds = input.getBoundingClientRect();
        var activeInputLeft = Math.max(compositionBounds.left, inputBounds.left);
        var boundedLeft = Math.max(screenBounds.left, Math.min(activeInputLeft, screenBounds.right - 1));
        var maximumWidth = Math.max(screenBounds.right - boundedLeft, 1);
        composition.style.maxWidth = maximumWidth + "px";
        composition.style.overflow = "hidden";
        composition.style.direction = "rtl";
        input.style.width = Math.min(Math.max(inputBounds.width, 1), maximumWidth) + "px";
        input.style.maxWidth = maximumWidth + "px";
        input.style.overflow = "hidden";
        lockPageViewport();
    }

    function scheduleImeContainment() {
        containImeGeometry();
        window.setTimeout(containImeGeometry, 0);
        window.requestAnimationFrame(containImeGeometry);
    }

    function focusTerminal() {
        if (!input) {
            failPage();
            return;
        }
        input.focus({ preventScroll: true });
        scheduleImeContainment();
    }

    if (!input || !composition || !screen) {
        failPage();
        return;
    }
    ["compositionstart", "compositionend", "beforeinput", "input", "keydown", "focus"]
        .forEach(function (eventName) {
            input.addEventListener(eventName, scheduleImeContainment);
        });
    window.addEventListener("keydown", acceptHardwareKey, true);
    input.addEventListener("beforeinput", function (event) {
        if (event.isTrusted !== true) return;
        if (event.inputType === "insertText" || event.inputType === "insertFromComposition") {
            if (!event.isComposing && typeof event.data === "string" && event.data.length > 0) {
                sendText(event.data);
            }
            return;
        }
        if (event.inputType === "insertLineBreak" || event.inputType === "insertParagraph") {
            event.preventDefault();
            sendKey("enter", false, false, false);
            return;
        }
        if (event.inputType === "deleteContentBackward") {
            event.preventDefault();
            sendKey("backspace", false, false, false);
            return;
        }
        if (event.inputType !== "insertCompositionText" &&
            event.inputType !== "insertFromPaste") {
            terminalStatus.textContent = "Input unavailable with this terminal";
        }
    }, true);
    input.addEventListener("compositionstart", function () {
        compositionActive = true;
    }, true);
    input.addEventListener("compositionend", function () {
        compositionActive = false;
    }, true);
    input.addEventListener("compositionupdate", function (event) {
        if (typeof event.data === "string") {
            composition.textContent = "\u200e" + event.data + "\u200e";
        }
        scheduleImeContainment();
    });
    input.addEventListener("paste", function (event) {
        event.preventDefault();
        event.stopImmediatePropagation();
        var sanitized = sanitizePaste(event.clipboardData ? event.clipboardData.getData("text/plain") : "");
        if (sanitized === null) {
            terminalStatus.textContent = "Paste exceeds the terminal input limit";
        } else {
            pasteInput(sanitized);
        }
    }, true);

    new ResizeObserver(scheduleFit).observe(terminalHost);
    window.addEventListener("resize", scheduleFit);
    window.addEventListener("orientationchange", scheduleFit);
    window.addEventListener("scroll", lockPageViewport);
    if (window.visualViewport) {
        window.visualViewport.addEventListener("resize", scheduleFit);
        window.visualViewport.addEventListener("scroll", lockPageViewport);
    }

    function settleFonts() {
        fontsSettled = true;
        // xterm caches the cell size measured at open(), before the vendored
        // face can arrive, and re-measures only on a font option change.
        terminal._core._charSizeService.measure();
        scheduleFit();
    }

    // Geometry stays local until the vendored faces settle, so the grid is
    // measured in the real font; a rejected load degrades to monospace rather
    // than withholding geometry.
    Promise.all([
        document.fonts.load('14px "JetBrains Mono"'),
        document.fonts.load('bold 14px "JetBrains Mono"')
    ]).then(settleFonts, settleFonts);

    function acceptAccessory(key) {
        if (key === "Control") {
            setModifiers(modifiers.control === "Off" ? "Armed" : "Off", modifiers.alt);
            focusTerminal();
            return;
        }
        if (key === "Alt") {
            setModifiers(modifiers.control, modifiers.alt === "Off" ? "Armed" : "Off");
            focusTerminal();
            return;
        }
        if (key === "Home" || key === "End") return;
        if (key === "PageUp" || key === "PageDown") {
            if (modifiers.control === "Armed" || modifiers.alt === "Armed") return;
            send({
                kind: "Scroll", source: "PageKey",
                direction: key === "PageUp" ? "Up" : "Down", lines: terminal.rows
            });
            focusTerminal();
            return;
        }
        var keys = {
            Escape: "escape", Slash: "/", Hyphen: "-", Tab: "tab",
            Up: "up", Down: "down", Left: "left", Right: "right"
        };
        if (!Object.prototype.hasOwnProperty.call(keys, key)) {
            failPage();
            return;
        }
        sendKey(keys[key], false, false, false);
        focusTerminal();
    }

    function acceptNativeMessage(message) {
        if (pageFailed) return;
        var payload = parseObject(message.data);
        if (!payload) {
            failPage();
            return;
        }
        if (payload.kind === "Frame" &&
            exactObject(payload, ["kind", "seq", "columns", "rows", "full", "ansiBase64"])) {
            var validSequence = typeof payload.seq === "string" &&
                /^[1-9][0-9]*$/.test(payload.seq) &&
                (payload.seq.length < 20 || payload.seq.length === 20 &&
                    payload.seq <= "18446744073709551615");
            var validGeometry = Number.isInteger(payload.columns) &&
                payload.columns >= minimumColumns && payload.columns <= maximumColumns &&
                Number.isInteger(payload.rows) &&
                payload.rows >= minimumRows && payload.rows <= maximumRows;
            var bytes = decodeBase64(payload.ansiBase64);
            if (!validSequence || !validGeometry || typeof payload.full !== "boolean" ||
                bytes === null || !payload.full &&
                (terminal.cols !== payload.columns || terminal.rows !== payload.rows)) {
                failPage();
                return;
            }
            if (payload.full) {
                if (terminalTouchInteraction) terminalTouchInteraction.cancel();
                terminal.reset();
                terminal.resize(payload.columns, payload.rows);
            }
            terminal.write(bytes, function () {
                terminalStatus.textContent = "Terminal connected";
                send({ kind: "OutputApplied", sequence: payload.seq });
            });
            return;
        }
        if (payload.kind === "Focus" && exactObject(payload, ["kind"])) {
            focusTerminal();
            return;
        }
        if (payload.kind === "Accessory" && exactObject(payload, ["kind", "key"]) &&
            typeof payload.key === "string") {
            acceptAccessory(payload.key);
            return;
        }
        if (payload.kind === "ResetInputState" && exactObject(payload, ["kind"])) {
            resetInputState();
            return;
        }
        if (payload.kind === "FontSize" && exactObject(payload, ["kind", "fontSizeCssPx"]) &&
            isFontSizeCssPx(payload.fontSizeCssPx)) {
            terminal.options.fontSize = payload.fontSizeCssPx;
            scheduleFit();
            return;
        }
        if (payload.kind === "Scroll" && exactObject(payload, ["kind", "direction"]) &&
            (payload.direction === "Backward" || payload.direction === "Forward")) {
            terminalTouchInteraction.scroll(payload.direction);
            return;
        }
        if (payload.kind === "ClearSelection" && exactObject(payload, ["kind", "generation"]) &&
            typeof payload.generation === "string") {
            terminalTouchInteraction.clearSelection(payload.generation);
            return;
        }
        failPage();
    }

    function acceptPagePort(event) {
        var handshake = parseObject(event.data);
        var validHandshake = event.ports && event.ports.length === 1 &&
            handshake &&
            exactObject(handshake, ["kind", "version", "longPressMilliseconds", "fontSizeCssPx"]) &&
            handshake.kind === "PagePort" && handshake.version === 4 &&
            typeof handshake.longPressMilliseconds === "number" &&
            Number.isFinite(handshake.longPressMilliseconds) &&
            Number.isInteger(handshake.longPressMilliseconds) &&
            handshake.longPressMilliseconds > 0 &&
            isFontSizeCssPx(handshake.fontSizeCssPx);
        if (!validHandshake) {
            failPage();
            return;
        }
        if (pagePort !== null) {
            failPage();
            return;
        }
        pagePort = event.ports[0];
        if (pageFailed) {
            send({ kind: "PageFailure" });
            pagePort = null;
            return;
        }
        terminal.options.fontSize = handshake.fontSizeCssPx;
        terminalTouchInteraction = createTerminalTouchInteraction({
            terminal: terminal,
            screen: screen,
            longPressMilliseconds: handshake.longPressMilliseconds
        });
        pagePort.onmessage = acceptNativeMessage;
        pagePort.onmessageerror = failPage;
        pagePort.start();
        publishModifiers();
        send({ kind: "Ready" });
        scheduleFit();
    }
}());
