/* common.js — updated features */
(function () {
  "use strict";

  /* ── Theme ─────────────────────────────────────────────────────────── */
  var THEME_KEY = "theme";

  function getPreferred() {
    var saved = localStorage.getItem(THEME_KEY);
    if (saved) return saved;
    return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
  }

  function applyTheme(theme) {
    if (theme === "dark") {
      document.documentElement.setAttribute("data-theme", "dark");
    } else {
      document.documentElement.removeAttribute("data-theme");
    }

    if (document.readyState === "interactive" || document.readyState === "complete") {
      syncCodeTheme(theme);
    }
  }

  /* ── Dynamic Code Highlight Color Script ───────────────────────────── */
  function syncCodeTheme(theme) {
    document.querySelectorAll('.content pre').forEach(function (pre) {
      if (theme === 'dark') {
        if (!pre.dataset.origBg) pre.dataset.origBg = pre.style.backgroundColor || "";
        pre.style.backgroundColor = 'var(--code-bg)';

        pre.querySelectorAll('span[style*="color"]').forEach(function (span) {
          span.style.filter = 'invert(1) hue-rotate(180deg) brightness(1.3) contrast(1.1)';
        });
      } else {
        if (pre.dataset.origBg !== undefined) pre.style.backgroundColor = pre.dataset.origBg;
        pre.querySelectorAll('span[style*="color"]').forEach(function (span) {
          span.style.filter = '';
        });
      }
    });
  }

  applyTheme(getPreferred());

  /* ── Sidebar state ──────────────────────────────────────────────────── */
  var SIDEBAR_KEY = "sidebar";

  function getSidebarState() {
    return localStorage.getItem(SIDEBAR_KEY) !== "hidden";
  }

  function applySidebar(visible) {
    if (visible) {
      document.body.classList.remove("sidebar-hidden");
    } else {
      document.body.classList.add("sidebar-hidden");
    }
  }

  /* ── Build TOC & Mobile Nav in sidebar ──────────────────────────────── */
  function buildSidebarTOC() {
    var sidebar = document.querySelector(".sidebar");
    if (!sidebar) return;

    // 1. Migrate Top Nav for Mobile
    var mainNav = document.getElementById("main-nav");
    if (mainNav && !sidebar.querySelector(".mobile-main-nav")) {
      var mobileNav = document.createElement("nav");
      mobileNav.className = "mobile-main-nav";
      mobileNav.innerHTML = mainNav.innerHTML;
      sidebar.insertBefore(mobileNav, sidebar.firstChild);
    }

    // 2. Scrape and parse the ToC specific to your parser's output
    var tocHeading = document.getElementById("table-of-contents");
    var tocList = null;
    var ulToRemove = null;

    if (tocHeading && tocHeading.nextElementSibling && tocHeading.nextElementSibling.tagName === "UL") {
      ulToRemove = tocHeading.nextElementSibling;
      tocList = parseTOCFromUL(ulToRemove);
    } else {
      tocList = buildTOCFromHeadings(); // Fallback
    }

    if (!tocList || !tocList.children.length) return;

    var label = document.createElement("div");
    label.className = "sidebar-label";
    label.textContent = "On this page";

    // Insert after the mobile nav (if it exists)
    var mobileNavRef = sidebar.querySelector(".mobile-main-nav");
    if (mobileNavRef) {
      mobileNavRef.after(label);
      label.after(tocList);
    } else {
      sidebar.insertBefore(tocList, sidebar.firstChild);
      sidebar.insertBefore(label, sidebar.firstChild);
    }

    // 3. Delete the original ToC from the top of the document
    if (tocHeading) tocHeading.remove();
    if (ulToRemove) ulToRemove.remove();
  }

  function parseTOCFromUL(srcUL) {
    var ul = document.createElement("ul");
    ul.className = "sidebar-toc";

    function processItems(srcList, destList, depth) {
      var items = srcList.children;
      for (var i = 0; i < items.length; i++) {
        var item = items[i];
        if (item.tagName !== "LI") continue;

        var a = item.querySelector(":scope > a");
        if (!a) continue;

        var li = document.createElement("li");
        li.className = "toc-h" + (depth + 1);

        var newA = document.createElement("a");
        newA.href = a.getAttribute("href");
        newA.textContent = a.textContent;
        li.appendChild(newA);

        var nested = item.querySelector(":scope > ul");
        if (nested) {
          var subUL = document.createElement("ul");
          subUL.className = "sidebar-toc";
          processItems(nested, subUL, depth + 1);
          li.appendChild(subUL);
        }

        destList.appendChild(li);
      }
    }

    processItems(srcUL, ul, 1);
    return ul;
  }

  function buildTOCFromHeadings() {
    var headings = document.querySelectorAll(".content h1, .content h2, .content h3, .content h4");
    if (!headings.length) return null;

    var ul = document.createElement("ul");
    ul.className = "sidebar-toc";

    headings.forEach(function (h) {
      var level = parseInt(h.tagName[1], 10);
      var li = document.createElement("li");
      li.className = "toc-h" + level;

      var a = document.createElement("a");
      a.href = "#" + h.id;
      a.textContent = h.textContent;
      li.appendChild(a);
      ul.appendChild(li);
    });

    return ul;
  }

  /* ── Scrollspy ──────────────────────────────────────────────────────── */
  function initScrollspy() {
    var headings = Array.from(
      document.querySelectorAll(".content h1[id], .content h2[id], .content h3[id], .content h4[id]")
    );
    if (!headings.length) return;

    var tocLinks = document.querySelectorAll(".sidebar-toc a");

    function onScroll() {
      var scrollY = window.scrollY + 80;
      var current = headings[0];

      for (var i = 0; i < headings.length; i++) {
        if (headings[i].offsetTop <= scrollY) {
          current = headings[i];
        }
      }

      tocLinks.forEach(function (a) {
        a.classList.toggle("active", a.getAttribute("href") === "#" + current.id);
      });
    }

    window.addEventListener("scroll", onScroll, { passive: true });
    onScroll();
  }

  /* ── Math Repair & KaTeX ────────────────────────────────────────────── */
  function repairMangledMath() {
    // Finds <p> tags strictly containing block math that got parsed as Markdown
    document.querySelectorAll('.content p').forEach(function (p) {
      var text = p.textContent.trim();
      if (text.startsWith('$$') && text.endsWith('$$')) {
        var html = p.innerHTML;
        // Revert <em> back to underscores, <strong> to double underscores, <br> to newlines
        html = html.replace(/<br\s*\/?>/gi, '\n');
        html = html.replace(/<\/?em>/gi, '_');
        html = html.replace(/<\/?strong>/gi, '__');

        var tmp = document.createElement('div');
        tmp.innerHTML = html;
        p.textContent = tmp.textContent; // Set as clean, raw text for KaTeX to find
      }
    });
  }

  function renderMath() {
    repairMangledMath(); // Fix markdown destruction before running KaTeX

    var renderOptions = {
      delimiters: [
        { left: '$$', right: '$$', display: true },
        { left: '$', right: '$', display: false },
        { left: '\\(', right: '\\)', display: false },
        { left: '\\[', right: '\\]', display: true }
      ],
      throwOnError: false
    };

    if (typeof renderMathInElement === "undefined") {
      var KATEX = "https://cdn.jsdelivr.net/npm/katex@0.16.10/dist/";

      var link = document.createElement("link");
      link.rel = "stylesheet";
      link.href = KATEX + "katex.min.css";
      document.head.appendChild(link);

      var s1 = document.createElement("script");
      s1.src = KATEX + "katex.min.js";
      s1.onload = function () {
        var s2 = document.createElement("script");
        s2.src = KATEX + "contrib/auto-render.min.js";
        s2.onload = function () {
          renderMathInElement(document.body, renderOptions);
        };
        document.head.appendChild(s2);
      };
      document.head.appendChild(s1);
    } else {
      renderMathInElement(document.body, renderOptions);
    }
  }

  /* ── Active nav links ───────────────────────────────────────────────── */
  function markActive() {
    var path = window.location.pathname.replace(/\/$/, "") || "/";
    document.querySelectorAll(".navbar nav a[href], .mobile-main-nav a[href]").forEach(function (a) {
      var href = a.getAttribute("href").replace(/\/$/, "") || "/";
      a.classList.toggle("active", href === path);
    });
  }

  /* ── DOMContentLoaded ───────────────────────────────────────────────── */
  document.addEventListener("DOMContentLoaded", function () {

    // IMPORTANT FIX: Inject 'sidebar-wrap' class to the parent div so CSS can position it!
    var sidebarEl = document.querySelector(".sidebar");
    if (sidebarEl && sidebarEl.parentElement && sidebarEl.parentElement.tagName === "DIV") {
      sidebarEl.parentElement.classList.add("sidebar-wrap");
    }

    syncCodeTheme(getPreferred());

    document.querySelectorAll(".content img").forEach(function (img) {
      if (!img.getAttribute("loading")) {
        img.setAttribute("loading", "lazy");
      }
    });

    var themeBtn = document.getElementById("theme-toggle");
    if (themeBtn) {
      themeBtn.addEventListener("click", function () {
        var current = document.documentElement.getAttribute("data-theme") === "dark" ? "dark" : "light";
        var next = current === "dark" ? "light" : "dark";
        applyTheme(next);
        localStorage.setItem(THEME_KEY, next);
      });
    }

    /* --- Sidebar & Overlay logic --- */
    var sidebarBtn = document.getElementById("sidebar-toggle");
    var overlay = document.createElement('div');
    overlay.className = 'sidebar-overlay';
    document.body.appendChild(overlay);

    function closeMobileSidebar() {
      document.body.classList.remove('mobile-sidebar-open');
    }

    overlay.addEventListener('click', closeMobileSidebar);

    if (sidebarBtn) {
      if (window.innerWidth > 768) {
        applySidebar(getSidebarState());
      } else {
        document.body.classList.remove("sidebar-hidden");
      }

      sidebarBtn.addEventListener("click", function () {
        if (window.innerWidth > 768) {
          var isVisible = !document.body.classList.contains("sidebar-hidden");
          applySidebar(!isVisible);
          localStorage.setItem(SIDEBAR_KEY, isVisible ? "hidden" : "visible");
        } else {
          document.body.classList.toggle("mobile-sidebar-open");
        }
      });
    }

    if (sidebarEl) {
      sidebarEl.addEventListener("click", function (e) {
        if (window.innerWidth <= 768 && e.target.tagName === "A") {
          closeMobileSidebar();
        }
      });
    }

    window.addEventListener("resize", function () {
      if (window.innerWidth > 768) {
        closeMobileSidebar();
      }
    });

    document.querySelectorAll(".content p").forEach(function (p) {
      if (p.textContent.trim() === "{:toc}") p.remove();
    });

    buildSidebarTOC();
    initScrollspy();
    markActive();

    var rawTextContent = document.body.textContent || document.body.innerText;
    if (rawTextContent.includes("$$") || rawTextContent.includes("\\[") || rawTextContent.includes("\\(")) {
      renderMath();
    }
  });

})();
