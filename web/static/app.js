/* ColdCase Console — progressive enhancement.
   No inline handlers; everything is wired here so a strict CSP can be used. */
(function () {
  "use strict";

  var form = document.getElementById("run-form");
  if (form) setupRun(form);
  setupTerminalTools();

  /* Confirm destructive actions declared via data-confirm. */
  document.addEventListener("submit", function (e) {
    var msg = e.target.getAttribute && e.target.getAttribute("data-confirm");
    if (msg && !window.confirm(msg)) e.preventDefault();
  });

  function splitArgs(text) {
    return text
      .split("\n")
      .map(function (l) { return l.trim(); })
      .filter(function (l) { return l.length > 0; });
  }

  function setupTerminalTools() {
    var term = document.getElementById("terminal");
    var wrap = document.getElementById("wrap-toggle");
    var clear = document.getElementById("clear-btn");
    if (wrap && term) {
      wrap.addEventListener("click", function () {
        var on = term.style.whiteSpace !== "pre";
        term.style.whiteSpace = on ? "pre" : "pre-wrap";
        wrap.setAttribute("aria-pressed", String(on));
      });
    }
    if (clear && term) {
      clear.addEventListener("click", function () { term.textContent = ""; });
    }
  }

  function setupRun(form) {
    var term = document.getElementById("terminal");
    var statusEl = document.getElementById("run-status");
    var runBtn = document.getElementById("run-btn");
    var stopBtn = document.getElementById("stop-btn");
    var tool = form.dataset.tool;
    var needsRoot = form.dataset.needsroot === "true";
    var csrf = form.dataset.csrf;
    var controller = null;

    function clearTerminal() { term.textContent = ""; }
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
      if (needsRoot && confirmEl && !confirmEl.checked) {
        setStatus("Confirm the elevated-privilege warning before running.", "is-error");
        return;
      }

      var args = splitArgs(document.getElementById("args").value);
      var session = document.getElementById("session").value;

      clearTerminal();
      sys("$ " + tool + (args.length ? " " + args.join(" ") : ""));
      setStatus("Running…");
      runBtn.disabled = true;
      stopBtn.disabled = false;
      controller = new AbortController();
      var started = Date.now();

      fetch("/api/run", {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-CSRF-Token": csrf },
        body: JSON.stringify({ tool: tool, args: args, session: session, confirm: true }),
        signal: controller.signal,
      }).then(function (res) {
        if (!res.ok || !res.body) {
          return res.text().then(function (t) {
            throw new Error(t || "HTTP " + res.status);
          });
        }
        return readSSE(res.body, {
          meta: function (d) {
            sys("[meta] " + d.binary + " · " + d.mode + (d.needsRoot ? " · elevated" : "") +
                (d.session ? " · session " + d.session : " · no session"));
          },
          stdout: function (d) { append(d.text); },
          stderr: function (d) { append(d.text, "t-stderr"); },
          exit: function (d) {
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
    var h = handlers[event];
    if (h) h(parsed);
  }
})();
