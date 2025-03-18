document.addEventListener('DOMContentLoaded', function() {
    // 获取DOM元素
    const elements = {
        menuToggle: document.querySelector('.menu-toggle'),
        menuClose: document.querySelector('.menu-close'),
        navLinks: document.querySelector('.nav-links'),
        menuOverlay: document.querySelector('.menu-overlay')
    };

    // 菜单状态管理
    const menuState = {
        isOpen: false,
        toggle() {
            this.isOpen ? this.close() : this.open();
        },
        open() {
            this.isOpen = true;
            this.updateUI();
        },
        close() {
            this.isOpen = false;
            this.updateUI();
        },
        updateUI() {
            const { menuToggle, navLinks, menuOverlay } = elements;
            const method = this.isOpen ? 'add' : 'remove';

            [menuToggle, navLinks, menuOverlay].forEach(el => el.classList[method]('active'));
            document.body.style.overflow = this.isOpen ? 'hidden' : '';
        }
    };

    // 事件监听
    function setupEventListeners() {
        const { menuToggle, menuClose, navLinks, menuOverlay } = elements;

        // 菜单切换
        menuToggle.addEventListener('click', () => menuState.toggle());
        menuClose.addEventListener('click', () => menuState.close());
        menuOverlay.addEventListener('click', () => menuState.close());

        // 导航链接点击（使用事件委托）
        navLinks.addEventListener('click', (event) => {
            if (event.target.tagName === 'A') {
                menuState.close();
            }
        });

        // 窗口大小改变
        window.addEventListener('resize', () => {
            if (window.innerWidth > 768 && menuState.isOpen) {
                menuState.close();
            }
        });
    }

    // 初始化
    setupEventListeners();
});

// 点击链接
function click_link(id) {
    fetch(`/links/${id}/click`, {
        method: 'POST',
    });
}

// 切换置顶状态
function togglePin(id) {
    fetch(`/links/${id}/pin`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
    })
    .then(response => response.json())
    .then(data => {
        if (data.error) {
            alert(data.error);
        } else {
            location.reload();
        }
    })
    .catch(error => {
        console.error('Error:', error);
        alert('操作失败');
    });
}