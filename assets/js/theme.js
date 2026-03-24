(function () {
  var root = document.documentElement;
  var toggle = document.getElementById('theme-toggle');

  function currentTheme() {
    return root.getAttribute('data-theme') || 'light';
  }

  function updateLabel() {
    if (!toggle) return;
    toggle.textContent = currentTheme() === 'dark' ? 'Light' : 'Dark';
  }

  function setTheme(theme) {
    root.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
    updateLabel();
  }

  if (toggle) {
    updateLabel();
    toggle.addEventListener('click', function () {
      setTheme(currentTheme() === 'dark' ? 'light' : 'dark');
    });
  }
})();
