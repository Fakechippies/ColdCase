/* ColdCase Console — progressive enhancement.
   No inline handlers; everything is wired here so a strict CSP can be used. */
(function () {
  "use strict";

  var runForm = document.getElementById("run-form");
  if (runForm) setupRun(runForm);
  setupTerminalTools();
  setupToolPicker();
  setupCategoryFilter();
  setupAvailabilityFilter();
  setupConfirmations();

  /* Confirm destructive actions declared via data-confirm. */
  function setupConfirmations() {
    document.addEventListener("submit", function (e) {
      var msg = e.target.getAttribute && e.target.getAttribute("data-confirm");
      if (msg && !window.confirm(msg)) e.preventDefault();
    });
  }

  function splitArgs(text) {
    return text
      .split("\n")
      .map(function (line) { return line.trim(); })
      .filter(function (line) { return line.length > 0; });
  }

  function setupToolPicker() {
    var picker = document.getElementById("tool-picker");
    var form = document.getElementById("tool-picker-form");
    if (picker && form) {
      picker.addEventListener("change", function () { form.submit(); });
    }
  }

  function setupTerminalTools() {
    var term = document.getElementById("terminal");
    var wrap = document.getElementById("wrap-toggle");
    var clear = document.getElementById("clear-btn");
    if (wrap && term) {
      wrap.addEventListener("click", function () {
        var nowrap = !term.classList.contains("is-nowrap");
        term.classList.toggle("is-nowrap", nowrap);
        wrap.setAttribute("aria-pressed", String(!nowrap));
      });
    }
    if (clear && term) {
      clear.addEventListener("click", function () { term.textContent = ""; });
    }
  }

  function setupCategoryFilter() {
    var chips = Array.prototype.slice.call(document.querySelectorAll("[data-filter]"));
    if (!chips.length) return;
    chips.forEach(function (chip) {
      chip.addEventListener("click", function () {
        var filter = chip.getAttribute("data-filter");
        chips.forEach(function (c) { c.classList.toggle("is-active", c === chip); });
        Array.prototype.forEach.call(document.querySelectorAll(".cat"), function (section) {
          section.hidden = filter !== "all" && section.getAttribute("data-category") !== filter;
        });
      });
    });
  }

  function setupAvailabilityFilter() {
    var chips = Array.prototype.slice.call(document.querySelectorAll("[data-avail-filter]"));
    var table = document.getElementById("avail-table");
    if (!chips.length || !table) return;

    var rows = Array.prototype.slice.call(table.querySelectorAll("tbody tr"));
    var search = document.getElementById("avail-search");
    var empty = document.getElementById("avail-empty");
    var active = "all";

    function apply() {
      var q = (search && search.value ? search.value : "").trim().toLowerCase();
      var shown = 0;
      rows.forEach(function (row) {
        var avail = row.getAttribute("data-avail") || "";
        var hay = (row.getAttribute("data-search") || "").toLowerCase();
        var ok = (active === "all" || avail === active) && (!q || hay.indexOf(q) >= 0);
        row.hidden = !ok;
        if (ok) shown++;
      });
      if (empty) empty.hidden = shown !== 0;
    }

    chips.forEach(function (chip) {
      chip.addEventListener("click", function () {
        active = chip.getAttribute("data-avail-filter") || "all";
        chips.forEach(function (c) { c.classList.toggle("is-active", c === chip); });
        apply();
      });
    });
    if (search) search.addEventListener("input", apply);
    apply();
  }

  function setupRun(form) {
    var term = document.getElementById("terminal");
    var statusEl = document.getElementById("run-status");
    var runBtn = document.getElementById("run-btn");
    var stopBtn = document.getElementById("stop-btn");
    var fileInput = document.getElementById("file");
    var fileName = document.getElementById("file-name");
    var dropzone = document.getElementById("dropzone");
    var tool = form.getAttribute("data-tool");
    var needsRoot = form.getAttribute("data-needsroot") === "true";
    var csrf = form.getAttribute("data-csrf");
    var controller = null;

    if (fileInput && fileName) {
      fileInput.addEventListener("change", function () {
        var file = fileInput.files && fileInput.files[0];
        fileName.textContent = file ? file.name : "Choose a file";
        fileName.title = file ? file.name : "";
      });
    }
    if (dropzone) {
      ["dragenter", "dragover"].forEach(function (evt) {
        dropzone.addEventListener(evt, function (e) {
          e.preventDefault();
          dropzone.classList.add("is-drag");
        });
      });
      ["dragleave", "drop"].forEach(function (evt) {
        dropzone.addEventListener(evt, function (e) {
          e.preventDefault();
          dropzone.classList.remove("is-drag");
        });
      });
    }

    function clearTerminal() { term.textContent = ""; }
    var sawOutput = false;
    function append(text, cls) {
      var span = document.createElement("span");
      if (cls) span.className = cls;
      span.textContent = text;
      term.appendChild(span);
      term.scrollTop = term.scrollHeight;
    }
    function sys(text, cls) {
      append(text.charAt(text.length - 1) === "\n" ? text : text + "\n", cls || "t-sys");
    }
    function setStatus(text, cls) {
      statusEl.textContent = text;
      statusEl.className = "run-status" + (cls ? " " + cls : "");
    }

    form.addEventListener("submit", function (e) {
      e.preventDefault();
      if (controller) return;

      var confirmEl = document.getElementById("confirm-root");
      var confirmed = !needsRoot || (confirmEl && confirmEl.checked);
      if (needsRoot && !confirmed) {
        setStatus("Confirm the elevated-privilege warning before running.", "is-error");
        return;
      }

      var args = splitArgs(document.getElementById("args").value);
      var session = document.getElementById("session").value;
      var file = fileInput && fileInput.files ? fileInput.files[0] : null;

      clearTerminal();
      sawOutput = false;
      var cmdline = tool + (args.length ? " " + args.join(" ") : "");
      if (file) cmdline += (args.length ? " " : " ") + file.name;
      sys("$ " + cmdline);
      setStatus("Running…");
      runBtn.disabled = true;
      stopBtn.disabled = false;
      controller = new AbortController();

      var body = new FormData();
      body.append("tool", tool);
      body.append("args", JSON.stringify(args));
      body.append("session", session);
      body.append("confirm", confirmed ? "1" : "0");
      if (file) body.append("file", file, file.name);

      fetch("/api/run", {
        method: "POST",
        headers: { "X-CSRF-Token": csrf },
        body: body,
        signal: controller.signal,
      }).then(function (res) {
        if (!res.ok || !res.body) {
          return res.text().then(function (text) {
            var message = text;
            try { message = JSON.parse(text).error || text; } catch (err) { /* keep text */ }
            throw new Error(message || "HTTP " + res.status);
          });
        }
        return readSSE(res.body, {
          meta: function (d) {
            var bits = [d.binary, d.mode];
            if (d.needsRoot) bits.push("elevated");
            bits.push(d.session ? "session " + d.session : "no session");
            if (d.uploaded) bits.push("upload stored");
            sys("[meta] " + bits.join(" · "));
          },
          stdout: function (d) { sawOutput = true; append(d.text); },
          stderr: function (d) { sawOutput = true; append(d.text, "t-stderr"); },
          exit: function (d) {
            if (d.code === 0 && !sawOutput) {
              sys("[no output] " + tool + " exited successfully without writing anything. " +
                  "This tool may not handle this kind of file — check its description or pick another tool.", "t-sys");
            }
            var bits = [];
            if (d.error) bits.push(d.error);
            if (d.logged) bits.push("logged to session " + d.logged);
            sys("[exit " + d.code + "]" + (bits.length ? " · " + bits.join(" · ") : ""),
                d.code === 0 ? "t-exit-0" : "t-exit-bad");
            setStatus(d.code === 0 ? "Completed" : "Exited with code " + d.code,
                      d.code === 0 ? "is-ok" : "is-error");
          },
          error: function (d) {
            sys("[error] " + d.message, "t-stderr");
            setStatus("Failed", "is-error");
          },
        });
      }).catch(function (err) {
        if (err.name === "AbortError") {
          sys("[stopped by user]", "t-stderr");
          setStatus("Stopped");
        } else {
          sys("[error] " + err.message, "t-stderr");
          setStatus("Failed", "is-error");
        }
      }).finally(function () {
        controller = null;
        runBtn.disabled = false;
        stopBtn.disabled = true;
      });
    });

    stopBtn.addEventListener("click", function () { if (controller) controller.abort(); });
  }

  function readSSE(body, handlers) {
    var reader = body.getReader();
    var decoder = new TextDecoder();
    var buf = "";
    function pump() {
      return reader.read().then(function (result) {
        if (result.done) {
          if (buf.trim()) dispatch(buf, handlers);
          return;
        }
        buf += decoder.decode(result.value, { stream: true });
        var idx;
        while ((idx = buf.indexOf("\n\n")) >= 0) {
          dispatch(buf.slice(0, idx), handlers);
          buf = buf.slice(idx + 2);
        }
        return pump();
      });
    }
    return pump();
  }

  function dispatch(frame, handlers) {
    var event = "message";
    var data = "";
    frame.split("\n").forEach(function (line) {
      if (line.indexOf("event:") === 0) event = line.slice(6).trim();
      else if (line.indexOf("data:") === 0) data += line.slice(5).replace(/^ /, "");
    });
    if (!data) return;
    var parsed;
    try { parsed = JSON.parse(data); } catch (e) { parsed = { text: data }; }
    var handler = handlers[event];
    if (handler) handler(parsed);
  }
})();
