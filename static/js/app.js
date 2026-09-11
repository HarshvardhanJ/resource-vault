// NITC PYQ Archive — Minimal Vanilla Helpers
document.addEventListener("DOMContentLoaded", function () {
  // Focus quick-search on '/' keypress if not inside input
  document.addEventListener("keydown", function (e) {
    if (e.key === "/" && !["INPUT", "TEXTAREA", "SELECT"].includes(document.activeElement.tagName)) {
      const searchInput = document.querySelector('input[name="q"]');
      if (searchInput) {
        e.preventDefault();
        searchInput.focus();
        searchInput.select();
      }
    }
  });
});
