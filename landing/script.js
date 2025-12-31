/**
 * YFlow Landing Page - Interactions
 */

// Copy command to clipboard
function copyCommand() {
  const command = document.getElementById('heroCommand').textContent;
  navigator.clipboard.writeText(command).then(() => {
    showCopyFeedback(document.querySelector('.hero-command .copy-btn'));
  });
}

// Copy to clipboard helper
function copyToClipboard(text, button) {
  navigator.clipboard.writeText(text).then(() => {
    showCopyFeedback(button);
  });
}

// Show copy feedback animation
function showCopyFeedback(button) {
  const originalHTML = button.innerHTML;
  button.innerHTML = `
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <polyline points="20 6 9 17 4 12"/>
    </svg>
  `;
  button.style.color = '#22c55e';

  setTimeout(() => {
    button.innerHTML = originalHTML;
    button.style.color = '';
  }, 2000);
}

// Smooth scroll for navigation links
document.querySelectorAll('a[href^="#"]').forEach(anchor => {
  anchor.addEventListener('click', function(e) {
    e.preventDefault();
    const target = document.querySelector(this.getAttribute('href'));
    if (target) {
      target.scrollIntoView({
        behavior: 'smooth',
        block: 'start'
      });
    }
  });
});

// Intersection Observer for scroll animations
const observerOptions = {
  threshold: 0.1,
  rootMargin: '0px 0px -50px 0px'
};

const fadeInObserver = new IntersectionObserver((entries) => {
  entries.forEach(entry => {
    if (entry.isIntersecting) {
      entry.target.classList.add('visible');
    }
  });
}, observerOptions);

// Observe elements for fade-in animation
document.querySelectorAll('.feature-card, .arch-card, .timeline-item, .command-card').forEach(el => {
  el.style.opacity = '0';
  el.style.transform = 'translateY(20px)';
  el.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
  fadeInObserver.observe(el);
});

// Add visible class styles
const style = document.createElement('style');
style.textContent = `
  .visible {
    opacity: 1 !important;
    transform: translateY(0) !important;
  }
`;
document.head.appendChild(style);

// Navbar background on scroll
const nav = document.querySelector('.nav');
let lastScroll = 0;

window.addEventListener('scroll', () => {
  const currentScroll = window.pageYOffset;

  if (currentScroll > 100) {
    nav.style.background = 'rgba(10, 10, 15, 0.95)';
    nav.style.backdropFilter = 'blur(20px)';
  } else {
    nav.style.background = 'rgba(10, 10, 15, 0.8)';
  }

  lastScroll = currentScroll;
});

// Animated counter for stats
function animateCounter(element, target, suffix = '') {
  const duration = 2000;
  const start = 0;
  const startTime = performance.now();

  function update(currentTime) {
    const elapsed = currentTime - startTime;
    const progress = Math.min(elapsed / duration, 1);
    const easeOutQuart = 1 - Math.pow(1 - progress, 4);
    const current = Math.floor(start + (target - start) * easeOutQuart);

    element.textContent = current + suffix;

    if (progress < 1) {
      requestAnimationFrame(update);
    }
  }

  requestAnimationFrame(update);
}

// CLI Demo Animation
function runCLIDemo() {
  const steps = [
    { id: 'cli-step-1', delay: 500 },
    { id: 'cli-step-2', delay: 2000 },
    { id: 'cli-step-3', delay: 4000 }
  ];

  steps.forEach(step => {
    setTimeout(() => {
      const element = document.getElementById(step.id);
      if (element) {
        element.style.opacity = '1';
        element.style.animation = 'fadeIn 0.5s ease forwards';
      }
    }, step.delay);
  });

  // Show UI changes after CLI demo
  setTimeout(() => {
    const uiRows = document.querySelectorAll('.ui-row:not(.header-row)');
    uiRows.forEach((row, index) => {
      setTimeout(() => {
        row.classList.add('updated');
      }, index * 300);
    });

    // Show change indicator
    setTimeout(() => {
      const indicator = document.getElementById('changeIndicator');
      if (indicator) {
        indicator.style.opacity = '1';
        indicator.style.animation = 'fadeIn 0.5s ease forwards';
      }
    }, 1500);
  }, 4500);
}

// Initialize CLI demo on page load
document.addEventListener('DOMContentLoaded', () => {
  // Reset CLI steps
  document.querySelectorAll('.cli-step').forEach(step => {
    step.style.opacity = '0';
  });

  // Start demo after a short delay
  setTimeout(runCLIDemo, 1000);

  // Animate arch stats when they come into view
  const archCards = document.querySelectorAll('.arch-card');
  archCards.forEach(card => {
    card.addEventListener('mouseenter', () => {
      const stats = card.querySelectorAll('.arch-stats span');
      stats.forEach(stat => {
        const text = stat.textContent;
        const numberMatch = text.match(/(\d+)/);
        if (numberMatch) {
          stat.style.transform = 'scale(1.05)';
          stat.style.transition = 'transform 0.3s ease';
          setTimeout(() => {
            stat.style.transform = 'scale(1)';
          }, 300);
        }
      });
    });
  });
});

// Parallax effect for gradient orbs
document.addEventListener('mousemove', (e) => {
  const orbs = document.querySelectorAll('.gradient-orb');
  const mouseX = e.clientX / window.innerWidth - 0.5;
  const mouseY = e.clientY / window.innerHeight - 0.5;

  orbs.forEach((orb, index) => {
    const speed = (index + 1) * 10;
    const x = mouseX * speed;
    const y = mouseY * speed;
    orb.style.transform = `translate(${x}px, ${y}px)`;
  });
});

// Keyboard navigation for roadmap
document.addEventListener('keydown', (e) => {
  if (e.key === 'Tab') {
    const timelineItems = document.querySelectorAll('.timeline-item');
    const focusableItems = Array.from(timelineItems).filter(item =>
      item.getBoundingClientRect().top < window.innerHeight
    );

    if (focusableItems.length > 0) {
      focusableItems[0].setAttribute('tabindex', '-1');
      focusableItems[0].focus();
    }
  }
});

// Add hover effects to feature cards
document.querySelectorAll('.feature-card').forEach(card => {
  card.addEventListener('mouseenter', () => {
    const icon = card.querySelector('.feature-icon');
    if (icon) {
      icon.style.transform = 'scale(1.1)';
      icon.style.transition = 'transform 0.3s ease';
    }
  });

  card.addEventListener('mouseleave', () => {
    const icon = card.querySelector('.feature-icon');
    if (icon) {
      icon.style.transform = 'scale(1)';
    }
  });
});

// Service worker registration for offline support (optional)
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    // Could register service worker here if needed
  });
}

// Console welcome message
console.log('%c🚀 YFlow - 强大的自托管 i18n 解决方案', 'color: #6366f1; font-size: 16px; font-weight: bold;');
console.log('%cGitHub: https://github.com/your-repo/yflow', 'color: #8b5cf6; font-size: 12px;');

// ============================================
// Internationalization (i18n)
// ============================================

let currentLang = localStorage.getItem('lang') || 'en';

// Initialize language
function initI18n() {
  const langSwitcher = document.querySelector('.lang-switcher');
  if (langSwitcher) {
    // Set initial active state
    updateLangButtons();

    // Add click handlers
    langSwitcher.querySelectorAll('.lang-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        const lang = btn.dataset.lang;
        if (lang !== currentLang) {
          setLanguage(lang);
        }
      });
    });
  }

  // Apply saved language
  setLanguage(currentLang, false);
}

// Update language button states
function updateLangButtons() {
  document.querySelectorAll('.lang-btn').forEach(btn => {
    if (btn.dataset.lang === currentLang) {
      btn.classList.add('active');
      btn.style.color = 'var(--text-primary)';
    } else {
      btn.classList.remove('active');
      btn.style.color = 'var(--text-muted)';
    }
  });
}

// Set language and update all translations
function setLanguage(lang, save = true) {
  currentLang = lang;

  if (save) {
    localStorage.setItem('lang', lang);
  }

  // Update button states
  updateLangButtons();

  // Update document language
  document.documentElement.lang = lang === 'zh' ? 'zh-CN' : 'en';

  // Update all translatable elements
  document.querySelectorAll('[data-i18n]').forEach(element => {
    const key = element.dataset.i18n;
    if (window.translations && window.translations[lang] && window.translations[lang][key]) {
      element.innerHTML = window.translations[lang][key];
    }
  });

  // Update page title
  if (window.translations && window.translations[lang]) {
    const title = lang === 'zh'
      ? 'YFlow - 强大的自托管 i18n 解决方案'
      : 'YFlow - Powerful Self-Hosted i18n Solution';
    document.title = title;
  }

  console.log(`Language switched to: ${lang}`);
}

// Make setLanguage available globally
window.setLanguage = setLanguage;
window.getLanguage = () => currentLang;

// Initialize i18n when DOM is ready
document.addEventListener('DOMContentLoaded', initI18n);
