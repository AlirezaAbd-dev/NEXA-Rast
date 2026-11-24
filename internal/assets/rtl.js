(() => {
  if (window.__nexaRastInstalled) {
    return;
  }
  window.__nexaRastInstalled = true;

  const STYLE_ID = "nexa-rast-style";
  const CSS_TEXT = __NEXA_RAST_CSS__;
  const TARGET_SELECTOR =
    'p, h1, h2, h3, h4, h5, h6, li, blockquote, td, th, ul, ol, textarea, [contenteditable="true"]';
  const SKIP_SELECTOR = "pre, code, kbd, samp, .monaco-editor";

  function injectStyle() {
    if (document.getElementById(STYLE_ID)) {
      return;
    }
    const style = document.createElement("style");
    style.id = STYLE_ID;
    style.textContent = CSS_TEXT;
    (document.head || document.documentElement).appendChild(style);
  }

  function applyAutoDir(el) {
    if (!(el instanceof Element) || el.closest(SKIP_SELECTOR)) {
      return;
    }
    if (el.matches(TARGET_SELECTOR) && !el.hasAttribute("dir")) {
      el.setAttribute("dir", "auto");
    }
    const descendants = el.querySelectorAll(TARGET_SELECTOR);
    for (let i = 0; i < descendants.length; i++) {
      const child = descendants[i];
      if (!child.hasAttribute("dir") && !child.closest(SKIP_SELECTOR)) {
        child.setAttribute("dir", "auto");
      }
    }
  }

  const pending = new Set();
  let rafScheduled = false;

  function flushPending() {
    rafScheduled = false;
    pending.forEach((node) => applyAutoDir(node));
    pending.clear();
  }

  function scheduleNode(node) {
    if (!(node instanceof Element)) {
      return;
    }
    pending.add(node);
    if (!rafScheduled) {
      rafScheduled = true;
      requestAnimationFrame(flushPending);
    }
  }

  function start() {
    injectStyle();
    if (document.body) {
      applyAutoDir(document.body);
    }

    const observer = new MutationObserver((mutations) => {
      for (let i = 0; i < mutations.length; i++) {
        const added = mutations[i].addedNodes;
        for (let j = 0; j < added.length; j++) {
          scheduleNode(added[j]);
        }
      }
    });

    const root = document.body || document.documentElement;
    if (root) {
      observer.observe(root, { childList: true, subtree: true });
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", start, { once: true });
  } else {
    start();
  }
})();
