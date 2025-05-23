// 更新题目序号
function updateQuestionNumbers() {
    const questionForms = document.querySelectorAll('.question-form');
    questionForms.forEach((form, index) => {
        const questionNumber = index + 1;
        const label = form.querySelector('label[for^="question-"]');
        if (label) {
            label.textContent = `题目 ${questionNumber}`;
        }
    });
}

// 创建考试
function createExam() {
    if (!validateExamForm()) {
        return;
    }

    const examData = collectExamFormData();

    fetch('/api/teachers/exams', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(examData),
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('创建考试失败');
            }
            return response.json();
        })
        .then(data => {
            alert('考试创建成功！');
            showDashboard();
        })
        .catch(error => {
            console.error('创建考试失败:', error);
            alert('创建考试失败，请重试');
        });
}

// 更新考试
function updateExam() {
    if (!validateExamForm()) {
        return;
    }

    const examData = collectExamFormData();

    fetch(`/api/teachers/exams/${currentExam.id}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(examData),
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('更新考试失败');
            }
            return response.json();
        })
        .then(data => {
            alert('考试更新成功！');
            showDashboard();
        })
        .catch(error => {
            console.error('更新考试失败:', error);
            alert('更新考试失败，请重试');
        });
}

// 验证考试表单
function validateExamForm() {
    const title = document.getElementById('exam-title').value.trim();
    const startTime = document.getElementById('exam-start-time').value;
    const endTime = document.getElementById('exam-end-time').value;
    const duration = document.getElementById('exam-duration').value;

    if (!title || !startTime || !endTime || !duration) {
        alert('请填写所有考试基本信息字段');
        return false;
    }

    if (new Date(endTime) <= new Date(startTime)) {
        alert('结束时间必须晚于开始时间');
        return false;
    }

    const questionForms = document.querySelectorAll('.question-form');
    if (questionForms.length === 0) {
        alert('请至少添加一个题目');
        return false;
    }

    for (let i = 0; i < questionForms.length; i++) {
        const questionForm = questionForms[i];
        const content = questionForm.querySelector('.question-content').value.trim();
        const type = questionForm.querySelector('.question-type').value;
        const score = questionForm.querySelector('.question-score').value;

        if (!content || !score) {
            alert(`请完善题目 ${i + 1} 的内容和分值`);
            return false;
        }

        if (type !== 'text') {
            const options = questionForm.querySelectorAll('.option-content');
            if (options.length < 2) {
                alert(`题目 ${i + 1} 至少需要两个选项`);
                return false;
            }

            for (let j = 0; j < options.length; j++) {
                if (!options[j].value.trim()) {
                    alert(`请完善题目 ${i + 1} 的选项 ${j + 1}`);
                    return false;
                }
            }

            const answer = questionForm.querySelector('.question-answer').value.trim();
            if (!answer) {
                alert(`请提供题目 ${i + 1} 的正确答案`);
                return false;
            }
        }
    }

    return true;
}

// 收集考试表单数据
function collectExamFormData() {
    const title = document.getElementById('exam-title').value.trim();
    const startTime = document.getElementById('exam-start-time').value;
    const endTime = document.getElementById('exam-end-time').value;
    const duration = parseInt(document.getElementById('exam-duration').value);

    // 设置考试状态
    let status = 'upcoming';
    const now = new Date();
    const examStartTime = new Date(startTime);
    const examEndTime = new Date(endTime);

    if (now > examEndTime) {
        status = 'ended';
    } else if (now >= examStartTime) {
        status = 'active';
    }

    // 创建考试对象
    const exam = {
        title: title,
        startTime: startTime,
        endTime: endTime,
        duration: duration,
        status: status
    };

    if (currentExam) {
        exam.id = currentExam.id;
    }

    // 收集题目数据
    const questions = [];
    const questionForms = document.querySelectorAll('.question-form');

    questionForms.forEach((form, index) => {
        const content = form.querySelector('.question-content').value.trim();
        const type = form.querySelector('.question-type').value;
        const score = parseInt(form.querySelector('.question-score').value);

        const question = {
            content: content,
            type: type,
            score: score
        };

        if (type !== 'text') {
            // 收集选项
            const options = [];
            form.querySelectorAll('.option-content').forEach(optionInput => {
                options.push(optionInput.value.trim());
            });
            question.options = options;

            // 收集答案
            question.answer = form.querySelector('.question-answer').value.trim();
        }

        questions.push(question);
    });

    return {
        exam: exam,
        questions: questions
    };
}
