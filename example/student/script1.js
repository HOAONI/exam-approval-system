document.addEventListener('DOMContentLoaded', function() {
    const navLinks = document.querySelectorAll('.nav-item');
    const pages = document.querySelectorAll('.page');
    const uploadPaperBtn = document.getElementById('uploadPaperBtn');
    const uploadPaperModal = document.getElementById('uploadPaperModal');
    const closeUploadModal = document.querySelector('#uploadPaperModal .close');
    const cancelUploadBtn = document.querySelector('#uploadPaperModal .cancel-btn');

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

    // 上传试卷模态框相关
    if (uploadPaperBtn) {
        uploadPaperBtn.addEventListener('click', function() {
            uploadPaperModal.style.display = 'block';
        });
    }

    if (closeUploadModal) {
        closeUploadModal.addEventListener('click', function() {
            uploadPaperModal.style.display = 'none';
        });
    }

    if (cancelUploadBtn) {
        cancelUploadBtn.addEventListener('click', function() {
            uploadPaperModal.style.display = 'none';
        });
    }

    // 点击模态框外部关闭
    window.addEventListener('click', function(event) {
        if (event.target == uploadPaperModal) {
            uploadPaperModal.style.display = 'none';
        }
    });

    // 修改密码模态框相关
    const changePasswordBtn = document.getElementById('changePasswordBtn');
    const changePasswordModal = document.getElementById('changePasswordModal');
    const closePasswordModal = document.querySelector('#changePasswordModal .close');
    const cancelPasswordBtn = document.querySelector('#changePasswordModal .cancel-btn');

    if (changePasswordBtn) {
        changePasswordBtn.addEventListener('click', function() {
            changePasswordModal.style.display = 'block';
        });
    }

    if (closePasswordModal) {
        closePasswordModal.addEventListener('click', function() {
            changePasswordModal.style.display = 'none';
        });
    }

    if (cancelPasswordBtn) {
        cancelPasswordBtn.addEventListener('click', function() {
            changePasswordModal.style.display = 'none';
        });
    }

    // 点击模态框外部关闭
    window.addEventListener('click', function(event) {
        if (event.target == changePasswordModal) {
            changePasswordModal.style.display = 'none';
        }
    });

    // 表单提交处理
    const uploadPaperForm = document.getElementById('uploadPaperForm');
    if (uploadPaperForm) {
        uploadPaperForm.addEventListener('submit', function(e) {
            e.preventDefault();
            // 这里添加上传试卷的逻辑
            alert('试卷提交成功！');
            uploadPaperModal.style.display = 'none';
            uploadPaperForm.reset();
        });
    }

    const changePasswordForm = document.getElementById('changePasswordForm');
    if (changePasswordForm) {
        changePasswordForm.addEventListener('submit', function(e) {
            e.preventDefault();
            // 这里添加修改密码的逻辑
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
});
