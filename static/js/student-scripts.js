// 检查登录状态
document.addEventListener('DOMContentLoaded', function() {
    const token = localStorage.getItem('token');
    const currentUser = JSON.parse(localStorage.getItem('currentUser') || '{}');
    
    // 如果没有token或者用户不是学生，重定向到登录页面
    if (!token || currentUser.role !== 'student') {
        window.location.href = '/login';
        return;
    }
    
    // 设置用户信息
    document.getElementById('user-name').textContent = currentUser.name || '未知用户';
    
    // 获取试卷数据
    fetchDashboardData();
});

// 获取仪表板数据
function fetchDashboardData() {
    const token = localStorage.getItem('token');
    
    fetch('/api/dashboard/student', {
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
    const rejectedPapersElement = document.querySelector('.card:nth-child(4) .card-info p');
    
    if (totalPapersElement) totalPapersElement.textContent = `${data.totalPapers || 0}份`;
    if (approvedPapersElement) approvedPapersElement.textContent = `${data.approvedPapers || 0}份`;
    if (pendingPapersElement) pendingPapersElement.textContent = `${data.pendingPapers || 0}份`;
    if (rejectedPapersElement) rejectedPapersElement.textContent = `${data.rejectedPapers || 0}份`;
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
            paper.status === 'approved' ? '已通过' :
            paper.status === 'pending' ? '审批中' :
            paper.status === 'rejected' ? '未通过' :
            paper.status === 'published' ? '已发布' : '草稿';
        
        const createdDate = new Date(paper.created_at).toLocaleDateString();
        
        tableContent += `
            <tr>
                <td>${paper.title}</td>
                <td>${paper.course}</td>
                <td>${createdDate}</td>
                <td>${paper.total_score > 0 ? paper.total_score : '--'}</td>
                <td><span class="status ${statusClass}">${statusText}</span></td>
            </tr>
        `;
    });
    
    tableBody.innerHTML = tableContent;
} 