(() => {
  const storeKey = 'lang';
  const defaultLang = 'zh';

  function currentLang() {
    const q = new URLSearchParams(location.search).get('lang');
    if (q) { localStorage.setItem(storeKey, q); return q; }
    return localStorage.getItem(storeKey) || defaultLang;
  }

  function applyI18n(dict) {
    document.querySelectorAll('[data-i18n]').forEach(el => {
      const k = el.getAttribute('data-i18n');
      if (dict[k] !== undefined) el.textContent = dict[k];
    });
    document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
      const k = el.getAttribute('data-i18n-placeholder');
      if (dict[k] !== undefined) el.setAttribute('placeholder', dict[k]);
    });
  }

  async function loadAndApply(lang) {
    try {
      const res = await fetch(`/static/i18n/${lang}.json`);
      const dict = await res.json();
      applyI18n(dict);
      const toggle = document.getElementById('lang-toggle');
      if (toggle) toggle.textContent = lang === 'zh' ? 'English' : '中文';
    } catch (e) { /* noop */ }
  }

  function initToggle() {
    const btn = document.getElementById('lang-toggle');
    if (!btn) return;
    btn.addEventListener('click', () => {
      const next = currentLang() === 'zh' ? 'en' : 'zh';
      localStorage.setItem(storeKey, next);
      loadAndApply(next);
    });
  }

  document.addEventListener('DOMContentLoaded', () => {
    initToggle();
    loadAndApply(currentLang());
  });
})();
