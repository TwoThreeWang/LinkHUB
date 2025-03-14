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

function click_link(linkId){
    // 创建一个新的XMLHttpRequest对象
    const xhr = new XMLHttpRequest();
    // 配置请求
    xhr.open('GET', `/links/${linkId}/click`, true);
    // 发送请求
    xhr.send();
}