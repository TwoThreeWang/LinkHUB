const {body} = document, themes = ['auto','light','dark'],
      toggle = id => document.getElementById(id);
    let currentTheme = localStorage.theme || 'auto';

const updateTheme = theme => {
  const isDark = theme === 'dark' || (theme === 'auto' && matchMedia('(prefers-color-scheme: dark)').matches);
  body.classList.toggle('dark_mode', isDark);
  if(toggle('theme-toggle')) toggle('theme-toggle').textContent = `☀︎ ${theme}`;
};

updateTheme(currentTheme);

toggle('theme-toggle')?.addEventListener('click', () => {
  currentTheme = themes[(themes.indexOf(currentTheme) + 1) % themes.length];
  updateTheme(currentTheme);
  localStorage.theme = currentTheme;
});

addEventListener('DOMContentLoaded', () => {
  matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => currentTheme === 'auto' && updateTheme('auto'));

  const topBtn = toggle('top');
  if (topBtn) addEventListener('scroll', () => topBtn.classList.toggle('show', scrollY > 200));

  window.showMsg = (msg, type = 'info') => {
    const el = Object.assign(document.createElement('div'), {className: `msg ${type}`, textContent: msg});
    document.body.append(el);
    requestAnimationFrame(() => el.classList.add('show'));
    setTimeout(() => el.classList.remove('show'), 3000);
  };
});

const lightbox = document.getElementById('lightbox');
const lightboxImg = document.getElementById('lightboxImg');

const openLightbox = img => {
    lightbox.style.display = 'block';
    lightboxImg.src = img.src;
    document.body.style.overflow = 'hidden';
};

const closeLightbox = () => {
    lightbox.style.display = 'none';
    document.body.style.overflow = 'auto';
};