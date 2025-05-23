/**
 * 试卷审批管理系统 JavaScript 主文件
 */

// 全局变量
const API_BASE_URL = '/api';
let token = localStorage.getItem('token');
let currentUser = null;

/**
 * 初始化应用
 */
document.addEventListener('DOMContentLoaded', function() {
    // 检查登录状态
    checkLoginStatus();
    
    // 更新用户信息
    updateUserInfo();
    
    // 注册退出登录按钮事件
    const logoutButton = document.getElementById('logout-button');
    if (logoutButton) {
        logoutButton.addEventListener('click', logout);
    }
    
    // 导航栏切换
    const navItems = document.querySelectorAll('.nav-item');
    const pages = document.querySelectorAll('.page');
    
    navItems.forEach(item => {
        item.addEventListener('click', function() {
            // 获取要显示的页面ID
            const pageId = this.getAttribute('data-page');
            
            // 移除所有导航项的激活状态
            navItems.forEach(nav => nav.classList.remove('active'));
            
            // 添加当前项的激活状态
            this.classList.add('active');
            
            // 隐藏所有页面
            pages.forEach(page => page.classList.remove('active'));
            
            // 显示目标页面
            const targetPage = document.getElementById(pageId);
            if (targetPage) {
                targetPage.classList.add('active');
            }
        });
    });
});

/**
 * 检查登录状态
 */
function checkLoginStatus() {
    // 从Cookie或本地存储中获取token
    const token = getCookie('token') || localStorage.getItem('token');
    
    if (!token) {
        // 如果没有token，重定向到登录页面
        window.location.href = '/login';
        return false;
    }
    
    // 验证token
    fetch('/api/auth/check', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        if (!response.ok) {
            // 如果token无效，清除token并重定向到登录页面
            clearToken();
            window.location.href = '/login';
        }
        return response.json();
    })
    .catch(() => {
        // 如果发生错误，清除token并重定向到登录页面
        clearToken();
        window.location.href = '/login';
    });
    
    return true;
}

/**
 * 退出登录
 */
function logout() {
    clearToken();
    window.location.href = '/login';
}

/**
 * 清除token
 */
function clearToken() {
    deleteCookie('token');
    localStorage.removeItem('token');
}

/**
 * 获取Cookie
 */
function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
}

/**
 * 删除Cookie
 */
function deleteCookie(name) {
    document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;`;
}

/**
 * 获取用户信息
 */
function getUserProfile() {
    const token = getCookie('token') || localStorage.getItem('token');
    
    if (!token) {
        return;
    }
    
    return fetch('/api/users/profile', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Failed to get user profile');
        }
        return response.json();
    });
}

/**
 * 更新UI显示用户信息
 */
function updateUserInfo() {
    getUserProfile().then(user => {
        if (user) {
            // 更新用户名
            const userNameElement = document.getElementById('user-name');
            if (userNameElement) {
                userNameElement.textContent = user.name;
            }
            
            // 其他UI更新...
        }
    }).catch(() => {
        // 处理错误
    });
}

/**
 * 初始化侧边栏切换
 */
function initSidebar() {
    const sidebarToggle = document.getElementById('sidebar-toggle');
    if (sidebarToggle) {
        sidebarToggle.addEventListener('click', function() {
            const sidebar = document.querySelector('.sidebar');
            sidebar.classList.toggle('show');
        });
    }
}

/**
 * 初始化表单处理
 */
function initFormHandlers() {
    // 登录表单处理
    const loginForm = document.getElementById('login-form');
    if (loginForm) {
        loginForm.addEventListener('submit', function(e) {
            e.preventDefault();
            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;
            
            // 检查是否有角色选择
            let role = '';
            const roleInput = document.getElementById('role');
            const roleRadios = document.querySelectorAll('input[name="role"]:checked');
            
            if (roleRadios.length > 0) {
                // 如果有角色单选按钮，使用选中的角色
                role = roleRadios[0].value;
            } else if (roleInput) {
                // 如果有隐藏的角色输入，使用它的值
                role = roleInput.value;
            }
            
            // 调用登录函数，角色参数可能为空
            login(username, password, role);
        });
    }

    // 创建考试表单处理
    const createExamForm = document.getElementById('create-exam-form');
    if (createExamForm) {
        createExamForm.addEventListener('submit', function(e) {
            e.preventDefault();
            const title = document.getElementById('title').value;
            const description = document.getElementById('description').value;
            const course = document.getElementById('course').value;
            const startTime = document.getElementById('start-time').value;
            const endTime = document.getElementById('end-time').value;
            
            createExam(title, description, course, startTime, endTime);
        });
    }

    // 创建试卷表单处理
    const createPaperForm = document.getElementById('create-paper-form');
    if (createPaperForm) {
        createPaperForm.addEventListener('submit', function(e) {
            e.preventDefault();
            const examId = document.getElementById('exam-id').value;
            const title = document.getElementById('title').value;
            const content = document.getElementById('content').value;
            const questions = document.getElementById('questions').value;
            const duration = document.getElementById('duration').value;
            const totalScore = document.getElementById('total-score').value;
            const passingScore = document.getElementById('passing-score').value;
            
            createPaper(examId, title, content, questions, duration, totalScore, passingScore);
        });
    }
}

/**
 * 用户登录
 * @param {string} username 用户名
 * @param {string} password 密码
 * @param {string} role 角色
 */
async function login(username, password, role) {
    showLoading('login-btn');
    
    try {
        const response = await fetch(`${API_BASE_URL}/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ username, password, role })
        });
        
        const data = await response.json();
        
        hideLoading('login-btn', '登录');
        
        if (response.ok) {
            // 登录成功
            // 保存令牌
            token = data.token;
            localStorage.setItem('token', token);
            
            // 保存用户信息
            currentUser = data.user;
            localStorage.setItem('currentUser', JSON.stringify(currentUser));
            
            // 更新UI
            updateUIForUser();
            
            // 重定向到控制面板
            window.location.href = '/dashboard';
        } else {
            // 登录失败，显示错误消息
            showAlert('login-alert', data.error || '登录失败', 'danger');
        }
    } catch (error) {
        hideLoading('login-btn', '登录');
        showAlert('login-alert', '登录失败，请检查网络连接', 'danger');
        console.error('登录错误:', error);
    }
}

/**
 * 创建考试
 * @param {string} title 考试标题
 * @param {string} description 考试描述
 * @param {string} course 课程名称
 * @param {string} startTime 开始时间
 * @param {string} endTime 结束时间
 */
async function createExam(title, description, course, startTime, endTime) {
    showLoading('create-exam-btn');
    
    try {
        const response = await fetch(`${API_BASE_URL}/exams`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({ title, description, course, start_time: startTime, end_time: endTime })
        });
        
        const data = await response.json();
        
        hideLoading('create-exam-btn', '创建考试');
        
        if (response.ok) {
            // 创建成功，显示成功消息
            showAlert('create-exam-alert', '考试创建成功', 'success');
            
            // 清空表单
            document.getElementById('create-exam-form').reset();
            
            // 2秒后刷新页面
            setTimeout(() => {
                window.location.href = '/dashboard?tab=my-exams';
            }, 2000);
        } else {
            // 创建失败，显示错误消息
            showAlert('create-exam-alert', data.error || '创建考试失败', 'danger');
        }
    } catch (error) {
        hideLoading('create-exam-btn', '创建考试');
        showAlert('create-exam-alert', '创建考试失败，请检查网络连接', 'danger');
        console.error('Create exam error:', error);
    }
}

/**
 * 创建试卷
 * @param {number} examId 考试ID
 * @param {string} title 试卷标题
 * @param {string} content 试卷内容
 * @param {string} questions 试题JSON字符串
 * @param {number} duration 考试时长（分钟）
 * @param {number} totalScore 总分
 * @param {number} passingScore 及格分数
 */
async function createPaper(examId, title, content, questions, duration, totalScore, passingScore) {
    showLoading('create-paper-btn');
    
    try {
        const response = await fetch(`${API_BASE_URL}/papers`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({
                exam_id: parseInt(examId),
                title,
                content,
                questions,
                duration: parseInt(duration),
                total_score: parseFloat(totalScore),
                passing_score: parseFloat(passingScore)
            })
        });
        
        const data = await response.json();
        
        hideLoading('create-paper-btn', '创建试卷');
        
        if (response.ok) {
            // 创建成功，显示成功消息
            showAlert('create-paper-alert', '试卷创建成功', 'success');
            
            // 清空表单
            document.getElementById('create-paper-form').reset();
            
            // 2秒后刷新页面
            setTimeout(() => {
                window.location.href = `/dashboard?tab=exam-detail&id=${examId}`;
            }, 2000);
        } else {
            // 创建失败，显示错误消息
            showAlert('create-paper-alert', data.error || '创建试卷失败', 'danger');
        }
    } catch (error) {
        hideLoading('create-paper-btn', '创建试卷');
        showAlert('create-paper-alert', '创建试卷失败，请检查网络连接', 'danger');
        console.error('Create paper error:', error);
    }
}

/**
 * 提交考试审批
 * @param {number} examId 考试ID
 */
async function submitExamForApproval(examId) {
    if (!confirm('确定要提交此考试进行审批吗？')) {
        return;
    }
    
    showLoading(`submit-exam-${examId}`);
    
    try {
        const response = await fetch(`${API_BASE_URL}/exams/${examId}/submit`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });
        
        const data = await response.json();
        
        hideLoading(`submit-exam-${examId}`, '提交审批');
        
        if (response.ok) {
            // 提交成功，刷新页面
            alert('考试已成功提交审批');
            window.location.reload();
        } else {
            alert(data.error || '提交审批失败');
        }
    } catch (error) {
        hideLoading(`submit-exam-${examId}`, '提交审批');
        alert('提交审批失败，请检查网络连接');
        console.error('Submit exam error:', error);
    }
}

/**
 * 审批考试
 * @param {number} examId 考试ID
 * @param {boolean} approved 是否批准
 * @param {string} comment 评论内容
 */
async function approveExam(examId, approved, comment) {
    const action = approved ? '批准' : '拒绝';
    if (!confirm(`确定要${action}此考试吗？`)) {
        return;
    }
    
    const buttonId = approved ? `approve-exam-${examId}` : `reject-exam-${examId}`;
    showLoading(buttonId);
    
    const url = approved ? 
        `${API_BASE_URL}/exams/${examId}/approve` : 
        `${API_BASE_URL}/exams/${examId}/reject`;
    
    try {
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({ comment })
        });
        
        const data = await response.json();
        
        hideLoading(buttonId, approved ? '批准' : '拒绝');
        
        if (response.ok) {
            // 操作成功，刷新页面
            alert(`考试已${action}`);
            window.location.reload();
        } else {
            alert(data.error || `${action}考试失败`);
        }
    } catch (error) {
        hideLoading(buttonId, approved ? '批准' : '拒绝');
        alert(`${action}考试失败，请检查网络连接`);
        console.error('Approve/reject exam error:', error);
    }
}

/**
 * 发布考试
 * @param {number} examId 考试ID
 */
async function publishExam(examId) {
    if (!confirm('确定要发布此考试吗？')) {
        return;
    }
    
    showLoading(`publish-exam-${examId}`);
    
    try {
        const response = await fetch(`${API_BASE_URL}/exams/${examId}/publish`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });
        
        const data = await response.json();
        
        hideLoading(`publish-exam-${examId}`, '发布');
        
        if (response.ok) {
            // 发布成功，刷新页面
            alert('考试已成功发布');
            window.location.reload();
        } else {
            alert(data.error || '发布考试失败');
        }
    } catch (error) {
        hideLoading(`publish-exam-${examId}`, '发布');
        alert('发布考试失败，请检查网络连接');
        console.error('Publish exam error:', error);
    }
}

/**
 * 添加评论
 * @param {number} examId 考试ID
 * @param {string} content 评论内容
 */
async function addComment(examId, content) {
    showLoading('add-comment-btn');
    
    try {
        const response = await fetch(`${API_BASE_URL}/exams/${examId}/comment`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({ content })
        });
        
        const data = await response.json();
        
        hideLoading('add-comment-btn', '提交评论');
        
        if (response.ok) {
            // 添加评论成功，刷新评论列表
            document.getElementById('comment-form').reset();
            alert('评论已成功提交');
            loadComments(examId);
        } else {
            alert(data.error || '提交评论失败');
        }
    } catch (error) {
        hideLoading('add-comment-btn', '提交评论');
        alert('提交评论失败，请检查网络连接');
        console.error('Add comment error:', error);
    }
}

/**
 * 加载考试评论
 * @param {number} examId 考试ID
 */
async function loadComments(examId) {
    const commentsContainer = document.getElementById('comments-container');
    if (!commentsContainer) return;
    
    commentsContainer.innerHTML = '<div class="text-center p-3"><div class="spinner"></div></div>';
    
    try {
        const response = await fetch(`${API_BASE_URL}/exams/${examId}/comments`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });
        
        if (response.ok) {
            const comments = await response.json();
            
            if (comments.length === 0) {
                commentsContainer.innerHTML = '<div class="p-3 text-center">暂无评论</div>';
                return;
            }
            
            // 渲染评论列表
            let html = '';
            comments.forEach(comment => {
                html += `
                <div class="card mb-3">
                    <div class="card-header">
                        <strong>${comment.user.name}</strong>
                        <span class="badge badge-${getRoleBadgeClass(comment.user.role)} ml-2">${getRoleDisplayName(comment.user.role)}</span>
                        <small class="float-right">${formatDate(comment.created_at)}</small>
                    </div>
                    <div class="card-body">
                        <p>${comment.content}</p>
                    </div>
                </div>
                `;
            });
            
            commentsContainer.innerHTML = html;
        } else {
            commentsContainer.innerHTML = '<div class="alert alert-danger">加载评论失败</div>';
        }
    } catch (error) {
        commentsContainer.innerHTML = '<div class="alert alert-danger">加载评论失败，请检查网络连接</div>';
        console.error('Load comments error:', error);
    }
}

/**
 * 获取角色徽章样式类
 * @param {string} role 角色
 * @returns {string} 徽章样式类
 */
function getRoleBadgeClass(role) {
    switch (role) {
        case 'teacher':
            return 'primary';
        case 'student':
            return 'success';
        case 'admin':
            return 'danger';
        default:
            return 'secondary';
    }
}

/**
 * 获取角色显示名称
 * @param {string} role 角色
 * @returns {string} 角色显示名称
 */
function getRoleDisplayName(role) {
    switch (role) {
        case 'teacher':
            return '教师';
        case 'student':
            return '学生';
        case 'admin':
            return '管理员';
        default:
            return '未知';
    }
}

/**
 * 格式化日期
 * @param {string} dateString 日期字符串
 * @returns {string} 格式化后的日期字符串
 */
function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
}

/**
 * 显示加载状态
 * @param {string} buttonId 按钮ID
 */
function showLoading(buttonId) {
    const button = document.getElementById(buttonId);
    if (button) {
        button.disabled = true;
        button.innerHTML = '<span class="spinner"></span> 处理中...';
    }
}

/**
 * 隐藏加载状态
 * @param {string} buttonId 按钮ID
 * @param {string} buttonText 按钮文本
 */
function hideLoading(buttonId, buttonText) {
    const button = document.getElementById(buttonId);
    if (button) {
        button.disabled = false;
        button.textContent = buttonText;
    }
}

/**
 * 显示提示信息
 * @param {string} alertId 提示框ID
 * @param {string} message 提示信息
 * @param {string} type 提示类型：success, danger, warning, info
 */
function showAlert(alertId, message, type = 'info') {
    const alertElement = document.getElementById(alertId);
    if (alertElement) {
        alertElement.className = `alert alert-${type}`;
        alertElement.textContent = message;
        alertElement.style.display = 'block';
        
        // 5秒后自动隐藏
        setTimeout(() => {
            alertElement.style.display = 'none';
        }, 5000);
    }
} 