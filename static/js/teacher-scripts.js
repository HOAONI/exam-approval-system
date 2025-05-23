// 检查登录状态
document.addEventListener('DOMContentLoaded', function() {
    const token = localStorage.getItem('token');
    const currentUser = JSON.parse(localStorage.getItem('currentUser') || '{}');
    
    // 如果没有token或者用户不是教师或管理员，重定向到登录页面
    if (!token || (currentUser.role !== 'teacher' && currentUser.role !== 'admin')) {
        window.location.href = '/login';
        return;
    }
    
    // 设置用户信息
    document.getElementById('user-name').textContent = currentUser.name || '未知用户';
    document.querySelector('.user').setAttribute('data-role', currentUser.role);
    
    // 获取仪表板数据
    fetchDashboardData();
});

// 获取仪表板数据
function fetchDashboardData() {
    const token = localStorage.getItem('token');
    const currentUser = JSON.parse(localStorage.getItem('currentUser') || '{}');
    
    // 根据角色选择不同的API
    const apiEndpoint = currentUser.role === 'admin' ? '/api/dashboard/admin' : '/api/dashboard/teacher';
    
    fetch(apiEndpoint, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('获取数据失败');
        }
        return response.json();
    })
    .then(data => {
        // 更新统计卡片
        updateStatsCards(data);
        
        // 更新最近试卷列表
        updateRecentPapers(data.recentPapers);
    })
    .catch(error => {
        console.error('获取仪表板数据失败:', error);
    });
}

// 更新统计卡片数据
function updateStatsCards(data) {
    const totalPapersElement = document.querySelector('.card:nth-child(1) .card-info p');
    const approvedPapersElement = document.querySelector('.card:nth-child(2) .card-info p');
    const pendingPapersElement = document.querySelector('.card:nth-child(3) .card-info p');
    const totalStudentsElement = document.querySelector('.card:nth-child(4) .card-info p');
    
    if (totalPapersElement) totalPapersElement.textContent = data.totalPapers || 0;
    if (approvedPapersElement) approvedPapersElement.textContent = data.approvedPapers || 0;
    if (pendingPapersElement) pendingPapersElement.textContent = data.pendingPapers || 0;
    if (totalStudentsElement) totalStudentsElement.textContent = data.totalStudents || 0;
}

// 更新最近试卷列表
function updateRecentPapers(papers) {
    const tableBody = document.querySelector('.recent-papers tbody');
    if (!tableBody) return;
    
    // 如果没有试卷数据
    if (!papers || papers.length === 0) {
        tableBody.innerHTML = '<tr><td colspan="5" class="text-center">暂无试卷数据</td></tr>';
        return;
    }
    
    // 生成表格行
    let tableContent = '';
    papers.forEach(paper => {
        const statusClass = 
            paper.status === 'approved' ? 'approved' :
            paper.status === 'pending' ? 'pending' :
            paper.status === 'rejected' ? 'rejected' :
            paper.status === 'published' ? 'published' : 'draft';
        
        const statusText = 
            paper.status === 'approved' ? '已批阅' :
            paper.status === 'pending' ? '待批阅' :
            paper.status === 'rejected' ? '已拒绝' :
            paper.status === 'published' ? '已发布' : '草稿';
        
        const createdDate = new Date(paper.created_at).toLocaleDateString('zh-CN', { 
            year: 'numeric', 
            month: '2-digit', 
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        });
        
        const creatorName = paper.creator ? paper.creator.name : '未知';
        
        const actionButton = paper.status === 'pending' 
            ? `<button class="action-btn" data-exam-id="${paper.id}">批阅</button>`
            : `<button class="action-btn view" data-exam-id="${paper.id}">查看</button>`;
        
        tableContent += `
            <tr>
                <td>${paper.title}</td>
                <td>${creatorName}</td>
                <td>${createdDate}</td>
                <td><span class="status ${statusClass}">${statusText}</span></td>
                <td>${actionButton}</td>
            </tr>
        `;
    });
    
    tableBody.innerHTML = tableContent;
    
    // 添加事件监听器
    document.querySelectorAll('.action-btn').forEach(button => {
        button.addEventListener('click', function() {
            const examId = this.getAttribute('data-exam-id');
            const action = this.classList.contains('view') ? '查看' : '批阅';
            alert(`${action}试卷ID: ${examId}`);
        });
    });
}

// 教师仪表板脚本
document.addEventListener('DOMContentLoaded', function() {
    console.log('Teacher scripts loaded');

    // 模态框相关
    const modalToggles = document.querySelectorAll('[data-toggle="modal"]');
    const modalBackdrop = document.getElementById('modalBackdrop');

    modalToggles.forEach(toggle => {
        toggle.addEventListener('click', function() {
            try {
                const targetModalId = this.getAttribute('data-target');
                const targetModal = document.getElementById(targetModalId);
                if (targetModal) {
                    targetModal.style.display = 'block';
                    modalBackdrop.style.display = 'block';
                    console.log('显示模态框:', targetModalId);
                } else {
                    console.error('未找到模态框:', targetModalId);
                }
            } catch (error) {
                console.error('模态框错误:', error);
            }
        });
    });

    window.hideModal = function(modalId) {
        try {
            const modal = document.getElementById(modalId);
            if (modal) {
                modal.style.display = 'none';
                modalBackdrop.style.display = 'none';
                console.log('隐藏模态框:', modalId);
            } else {
                console.error('未找到要隐藏的模态框:', modalId);
            }
        } catch (error) {
            console.error('隐藏模态框错误:', error);
        }
    };

    // 批阅试卷按钮点击事件
    const reviewButtons = document.querySelectorAll('.review-btn, [data-exam-id]');
    reviewButtons.forEach(button => {
        button.addEventListener('click', function() {
            try {
                const examId = this.getAttribute('data-exam-id');
                if (examId) {
                    alert('批阅试卷ID: ' + examId + ' (功能尚未实现)');
                    console.log('批阅试卷:', examId);
                } else {
                    console.error('未找到试卷ID');
                }
            } catch (error) {
                console.error('批阅试卷错误:', error);
            }
        });
    });

    // 审批按钮点击事件
    const approveButtons = document.querySelectorAll('.approve-btn');
    const rejectButtons = document.querySelectorAll('.reject-btn');
    
    approveButtons.forEach(button => {
        button.addEventListener('click', function() {
            try {
                const examId = this.getAttribute('data-exam-id');
                if (examId) {
                    alert('通过试卷ID: ' + examId + ' (功能尚未实现)');
                    console.log('通过试卷:', examId);
                } else {
                    console.error('未找到试卷ID');
                }
            } catch (error) {
                console.error('通过试卷错误:', error);
            }
        });
    });
    
    rejectButtons.forEach(button => {
        button.addEventListener('click', function() {
            try {
                const examId = this.getAttribute('data-exam-id');
                if (examId) {
                    alert('拒绝试卷ID: ' + examId + ' (功能尚未实现)');
                    console.log('拒绝试卷:', examId);
                } else {
                    console.error('未找到试卷ID');
                }
            } catch (error) {
                console.error('拒绝试卷错误:', error);
            }
        });
    });

    console.log('Teacher scripts initialization complete');
}); 