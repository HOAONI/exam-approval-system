document.addEventListener('DOMContentLoaded', function() {
    const navLinks = document.querySelectorAll('.nav-item');
    const pages = document.querySelectorAll('.page');

    // 模态框元素
    const createPaperBtn = document.getElementById('createPaperBtn');
    const createPaperModal = document.getElementById('createPaperModal');
    const importStudentsBtn = document.getElementById('importStudentsBtn');
    const importStudentsModal = document.getElementById('importStudentsModal');
    const changePasswordBtn = document.getElementById('changePasswordBtn');
    const changePasswordModal = document.getElementById('changePasswordModal');

    // 关闭按钮
    const closeButtons = document.querySelectorAll('.close');
    const cancelButtons = document.querySelectorAll('.cancel-btn');

    // 初始化：显示首页
    showPage('home');

    // 导航栏点击事件
    navLinks.forEach(link => {
        link.addEventListener('click', function() {
            const pageId = this.dataset.page;
            showPage(pageId);
        });
    });

    // 显示页面函数
    function showPage(pageId) {
        pages.forEach(page => {
            page.classList.remove('active');
        });
        navLinks.forEach(link => {
            link.classList.remove('active');
        });

        const selectedPage = document.getElementById(pageId);
        if (selectedPage) {
            selectedPage.classList.add('active');
            const selectedNavLink = document.querySelector(`.nav-item[data-page="${pageId}"]`);
            if (selectedNavLink) {
                selectedNavLink.classList.add('active');
            }
        }
    }

    // 打开模态框
    if (createPaperBtn) {
        createPaperBtn.addEventListener('click', function() {
            createPaperModal.style.display = 'block';
        });
    }

    if (importStudentsBtn) {
        importStudentsBtn.addEventListener('click', function() {
            importStudentsModal.style.display = 'block';
        });
    }

    if (changePasswordBtn) {
        changePasswordBtn.addEventListener('click', function() {
            changePasswordModal.style.display = 'block';
        });
    }

    // 关闭模态框
    closeButtons.forEach(button => {
        button.addEventListener('click', function() {
            const modal = this.closest('.modal');
            if (modal) {
                modal.style.display = 'none';
            }
        });
    });

    cancelButtons.forEach(button => {
        button.addEventListener('click', function() {
            const modal = this.closest('.modal');
            if (modal) {
                modal.style.display = 'none';
            }
        });
    });

    // 点击模态框外部关闭
    window.addEventListener('click', function(event) {
        if (event.target.classList.contains('modal')) {
            event.target.style.display = 'none';
        }
    });

    // 表单提交处理
    const createPaperForm = document.getElementById('createPaperForm');
    if (createPaperForm) {
        createPaperForm.addEventListener('submit', function(e) {
            e.preventDefault();
            // 这里添加创建试卷的逻辑
            alert('试卷创建成功！');
            createPaperModal.style.display = 'none';
            createPaperForm.reset();
        });
    }

    const importStudentsForm = document.getElementById('importStudentsForm');
    if (importStudentsForm) {
        importStudentsForm.addEventListener('submit', function(e) {
            e.preventDefault();
            // 这里添加导入学生的逻辑
            alert('学生名单导入成功！');
            importStudentsModal.style.display = 'none';
            importStudentsForm.reset();
        });
    }

    const changePasswordForm = document.getElementById('changePasswordForm');
    if (changePasswordForm) {
        changePasswordForm.addEventListener('submit', function(e) {
            e.preventDefault();
            // 验证密码
            const newPassword = document.getElementById('new-password').value;
            const confirmPassword = document.getElementById('confirm-password').value;

            if (newPassword !== confirmPassword) {
                alert('两次输入的密码不一致！');
                return;
            }

            alert('密码修改成功！');
            changePasswordModal.style.display = 'none';
            changePasswordForm.reset();
        });
    }

    // 保存个人信息
    const profileForm = document.querySelector('.profile-form');
    if (profileForm) {
        profileForm.addEventListener('submit', function(e) {
            e.preventDefault();
            // 这里添加保存个人信息的逻辑
            alert('个人信息保存成功！');
        });
    }

    // 下载模板按钮
    const templateBtn = document.querySelector('.template-btn');
    if (templateBtn) {
        templateBtn.addEventListener('click', function() {
            alert('模板下载中...');
            // 这里添加下载模板的逻辑
        });
    }
});
