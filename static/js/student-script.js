// static/student/script.js

// 全局变量
let currentUser = null;
let currentExam = null;
let examQuestions = [];
let answers = {};
let examTimer = null;
let remainingTime = 0;

// 页面加载时执行
document.addEventListener('DOMContentLoaded', function() {
    // 检查是否已登录
    const savedUser = localStorage.getItem('currentUser');
    if (savedUser) {
        currentUser = JSON.parse(savedUser);
        if (currentUser.role === 'student') {
            showDashboard();
        } else {
            // 如果不是学生角色，清除并重新登录
            localStorage.removeItem('currentUser');
            showLoginForm();
        }
    } else {
        showLoginForm();
    }

    // 登录表单提交
    document.getElementById('login-form').addEventListener('submit', function(e) {
        e.preventDefault();
        login();
    });

    // 退出登录
    document.getElementById('logout-btn').addEventListener('click', function() {
        logout();
    });
});

// 登录函数
function login() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    fetch('/api/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('登录失败');
            }
            return response.json();
        })
        .then(data => {
            if (data.role !== 'student') {
                throw new Error('无效的学生账号');
            }

            currentUser = data;
            localStorage.setItem('currentUser', JSON.stringify(currentUser));
            showDashboard();
        })
        .catch(error => {
            alert(error.message || '用户名或密码错误');
        });
}

// 登出函数
function logout() {
    localStorage.removeItem('currentUser');
    currentUser = null;
    showLoginForm();
}

// 显示登录表单
function showLoginForm() {
    document.getElementById('login-section').style.display = 'block';
    document.getElementById('dashboard-section').style.display = 'none';
    document.getElementById('exam-section').style.display = 'none';
}

// 显示学生仪表盘
function showDashboard() {
    document.getElementById('login-section').style.display = 'none';
    document.getElementById('dashboard-section').style.display = 'block';
    document.getElementById('exam-section').style.display = 'none';

    // 显示用户信息
    document.getElementById('student-name').textContent = currentUser.name;

    // 加载考试列表
    loadExams();
}

// 加载考试列表
function loadExams() {
    fetch(`/api/students/${currentUser.id}/exams`)
        .then(response => response.json())
        .then(data => {
            const upcomingContainer = document.getElementById('upcoming-exams');
            const pastContainer = document.getElementById('past-exams');

            upcomingContainer.innerHTML = '';
            pastContainer.innerHTML = '';

            data.forEach(exam => {
                const examCard = document.createElement('div');
                examCard.className = 'exam-card';

                let statusText, statusClass, actionButton = '';

                if (exam.status === 'not-started') {
                    statusText = '未开始';
                    statusClass = 'status-upcoming';
                    actionButton = `<button class="start-exam-btn" data-exam-id="${exam.examId}">开始考试</button>`;
                } else if (exam.status === 'in-progress') {
                    statusText = '进行中';
                    statusClass = 'status-active';
                    actionButton = `<button class="continue-exam-btn" data-exam-id="${exam.examId}">继续考试</button>`;
                } else if (exam.status === 'submitted') {
                    statusText = '已完成';
                    statusClass = 'status-completed';
                    actionButton = `<div class="exam-score">得分: ${exam.score !== null ? exam.score : '待评分'}</div>`;
                }

                const startTime = new Date(exam.startTime);
                const endTime = new Date(exam.endTime);

                examCard.innerHTML = `
                <h3>${exam.title}</h3>
                <div class="exam-info">
                    <p>开始时间: ${startTime.toLocaleString()}</p>
                    <p>结束时间: ${endTime.toLocaleString()}</p>
                    <p>时长: ${exam.duration} 分钟</p>
                    <p class="exam-status ${statusClass}">状态: ${statusText}</p>
                </div>
                <div class="exam-actions">
                    ${actionButton}
                </div>
            `;

                if (exam.status === 'submitted') {
                    pastContainer.appendChild(examCard);
                } else {
                    upcomingContainer.appendChild(examCard);
                }
            });

            // 添加开始考试按钮事件
            document.querySelectorAll('.start-exam-btn, .continue-exam-btn').forEach(btn => {
                btn.addEventListener('click', function() {
                    const examId = this.getAttribute('data-exam-id');
                    loadExam(examId);
                });
            });
        })
        .catch(error => {
            console.error('加载考试列表失败:', error);
            alert('加载考试列表失败，请重试');
        });
}

// 加载考试
function loadExam(examId) {
    fetch(`/api/students/exams/${examId}`)
        .then(response => response.json())
        .then(data => {
            currentExam = data.exam;
            examQuestions = data.questions;

            // 查看考试是否已经开始
            if (currentExam.status === 'not-started') {
                // 开始考试
                fetch(`/api/students/exams/${examId}/start`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ studentId: currentUser.id }),
                })
                    .then(response => response.json())
                    .then(() => {
                        showExam();
                    })
                    .catch(error => {
                        console.error('开始考试失败:', error);
                        alert('开始考试失败，请重试');
                    });
            } else {
                // 考试已经开始，直接显示
                showExam();
            }
        })
        .catch(error => {
            console.error('加载考试失败:', error);
            alert('加载考试失败，请重试');
        });
}

// 显示考试界面
function showExam() {
    document.getElementById('login-section').style.display = 'none';
    document.getElementById('dashboard-section').style.display = 'none';
    document.getElementById('exam-section').style.display = 'block';

    // 显示考试信息
    document.getElementById('exam-title').textContent = currentExam.title;

    // 初始化计时器
    startExamTimer(currentExam.duration);

    // 渲染题目
    renderQuestions();

    // 添加提交按钮事件
    document.getElementById('submit-exam-btn').addEventListener('click', submitExam);
}

// 渲染题目
function renderQuestions() {
    const questionsContainer = document.getElementById('questions-container');
    questionsContainer.innerHTML = '';

    examQuestions.forEach((question, index) => {
        const questionDiv = document.createElement('div');
        questionDiv.className = 'question';
        questionDiv.id = `question-${question.id}`;

        let questionContent = `
            <div class="question-header">
                <h3>问题 ${index + 1}</h3>
                <span class="question-score">${question.score} 分</span>
            </div>
            <div class="question-content">
                <p>${question.content}</p>
            </div>
        `;

        if (question.type === 'single') {
            questionContent += `<div class="options-container">`;
            question.options.forEach(option => {
                questionContent += `
                    <div class="option">
                        <input type="radio" name="question-${question.id}" value="${option}" id="option-${question.id}-${option}">
                        <label for="option-${question.id}-${option}">${option}</label>
                    </div>
                `;
            });
            questionContent += `</div>`;
        } else if (question.type === 'multiple') {
            questionContent += `<div class="options-container">`;
            question.options.forEach(option => {
                questionContent += `
                    <div class="option">
                        <input type="checkbox" name="question-${question.id}" value="${option}" id="option-${question.id}-${option}">
                        <label for="option-${question.id}-${option}">${option}</label>
                    </div>
                `;
            });
            questionContent += `</div>`;
        } else if (question.type === 'text') {
            questionContent += `
                <div class="text-answer">
                    <textarea id="answer-${question.id}" rows="6" placeholder="在此输入您的答案..."></textarea>
                </div>
            `;
        }

        questionDiv.innerHTML = questionContent;
        questionsContainer.appendChild(questionDiv);

        // 添加事件监听器以保存答案
        if (question.type === 'single') {
            const radios = questionDiv.querySelectorAll(`input[name="question-${question.id}"]`);
            radios.forEach(radio => {
                radio.addEventListener('change', function() {
                    answers[question.id] = this.value;
                });
            });
        } else if (question.type === 'multiple') {
            const checkboxes = questionDiv.querySelectorAll(`input[name="question-${question.id}"]`);
            checkboxes.forEach(checkbox => {
                checkbox.addEventListener('change', function() {
                    const selectedOptions = [];
                    checkboxes.forEach(cb => {
                        if (cb.checked) {
                            selectedOptions.push(cb.value);
                        }
                    });
                    answers[question.id] = selectedOptions.join(',');
                });
            });
        } else if (question.type === 'text') {
            const textarea = questionDiv.querySelector(`#answer-${question.id}`);
            textarea.addEventListener('input', function() {
                answers[question.id] = this.value;
            });
        }
    });
}

// 开始考试计时器
function startExamTimer(duration) {
    remainingTime = duration * 60; // 转换为秒

    const timerElement = document.getElementById('exam-timer');

    function updateTimer() {
        const hours = Math.floor(remainingTime / 3600);
        const minutes = Math.floor((remainingTime % 3600) / 60);
        const seconds = remainingTime % 60;

        timerElement.textContent = `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;

        if (remainingTime <= 0) {
            clearInterval(examTimer);
            submitExam();
        } else {
            remainingTime--;
        }
    }

    updateTimer();
    examTimer = setInterval(updateTimer, 1000);
}

// 提交考试
function submitExam() {
    if (!confirm('确定要提交考试吗？提交后将无法修改答案。')) {
        return;
    }

    clearInterval(examTimer);

    fetch(`/api/students/exams/${currentExam.id}/submit`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            studentId: currentUser.id,
            answers: answers
        }),
    })
        .then(response => response.json())
        .then(data => {
            alert(`考试已成功提交！\n您的得分: ${data.score !== null ? data.score : '待评分'}`);
            showDashboard();
        })
        .catch(error => {
            console.error('提交考试失败:', error);
            alert('提交考试失败，请重试');
        });
}
