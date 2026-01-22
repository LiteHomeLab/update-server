// 设置页面 JavaScript

// 页面切换功能
function nextStep(direction) {
    const steps = document.querySelectorAll('.step');
    const contents = document.querySelectorAll('.step-content');
    const currentStep = Array.from(contents).findIndex(content => content.classList.contains('active'));

    if (direction === -1 && currentStep > 0) {
        // 上一步
        steps[currentStep].classList.remove('active');
        contents[currentStep].classList.remove('active');
        steps[currentStep - 1].classList.add('active');
        contents[currentStep - 1].classList.add('active');
    } else if (direction === 1 && currentStep < 2) {
        // 下一步
        if (validateCurrentStep(currentStep + 1)) {
            steps[currentStep].classList.remove('active');
            contents[currentStep].classList.remove('active');
            steps[currentStep + 1].classList.add('active');
            contents[currentStep + 1].classList.add('active');

            // 如果是第二步到第三步，生成配置摘要
            if (currentStep + 1 === 2) {
                generateSummary();
            }
        }
    } else if (direction === 2 && currentStep === 1) {
        // 从第二步直接跳到第三步
        if (validateCurrentStep(2)) {
            steps[1].classList.remove('active');
            contents[1].classList.remove('active');
            steps[2].classList.add('active');
            contents[2].classList.add('active');
            generateSummary();
        }
    }
}

// 验证当前步骤
function validateCurrentStep(step) {
    const content = document.getElementById(`step${step}`);
    const form = content.querySelector('form');

    if (!form) return true;

    const requiredInputs = form.querySelectorAll('[required]');
    let isValid = true;

    for (const input of requiredInputs) {
        if (!input.value.trim()) {
            input.style.borderColor = '#e74c3c';
            isValid = false;

            // 添加错误提示
            let errorDiv = input.parentElement.querySelector('.error-message');
            if (!errorDiv) {
                errorDiv = document.createElement('div');
                errorDiv.className = 'error-message';
                errorDiv.textContent = '此字段为必填项';
                input.parentElement.appendChild(errorDiv);
            }
            errorDiv.style.display = 'block';
        } else {
            input.style.borderColor = '#ddd';

            // 移除错误提示
            const errorDiv = input.parentElement.querySelector('.error-message');
            if (errorDiv) {
                errorDiv.style.display = 'none';
            }
        }
    }

    // 特殊验证
    if (step === 2) {
        // 验证密码
        const password = document.getElementById('adminPassword').value;
        const confirmPassword = document.getElementById('confirmPassword').value;

        if (password && password !== confirmPassword) {
            document.getElementById('passwordError').style.display = 'block';
            isValid = false;
        } else {
            document.getElementById('passwordError').style.display = 'none';
        }

        // 验证密码强度
        if (password) {
            updatePasswordStrength(password);
        }
    }

    return isValid;
}

// 更新密码强度
function updatePasswordStrength(password) {
    const strengthBar = document.querySelector('.strength-bar');
    const strengthText = document.querySelector('.strength-text');

    let strength = 0;
    if (password.length >= 8) strength++;
    if (/[a-z]/.test(password) && /[A-Z]/.test(password)) strength++;
    if (/[0-9]/.test(password)) strength++;
    if (/[^A-Za-z0-9]/.test(password)) strength++;

    strengthBar.className = 'strength-bar';
    switch (strength) {
        case 0:
        case 1:
            strengthBar.classList.add('weak');
            strengthText.textContent = '密码强度：弱';
            break;
        case 2:
            strengthBar.classList.add('medium');
            strengthText.textContent = '密码强度：中';
            break;
        case 3:
        case 4:
            strengthBar.classList.add('strong');
            strengthText.textContent = '密码强度：强';
            break;
    }
}

// 生成配置摘要
function generateSummary() {
    const summary = document.getElementById('configSummary');

    const data = {
        '服务器名称': document.getElementById('serverName').value,
        '数据目录': document.getElementById('dataDir').value,
        '服务器端口': document.getElementById('port').value,
        '最大文件大小': document.getElementById('maxFileSize').value + ' MB',
        '管理员用户名': document.getElementById('adminUsername').value,
        '启用 HTTPS/TLS': document.getElementById('enableTLS').checked ? '是' : '否',
        '启用身份验证': document.getElementById('enableAuth').checked ? '是' : '否'
    };

    let summaryHTML = '';
    for (const [key, value] of Object.entries(data)) {
        summaryHTML += `
            <div class="summary-item">
                <span>${key}</span>
                <span>${value}</span>
            </div>
        `;
    }

    summary.innerHTML = summaryHTML;
}

// TLS 开关
document.getElementById('enableTLS').addEventListener('change', function() {
    const tlsGroup = document.getElementById('tlsGroup');
    tlsGroup.style.display = this.checked ? 'block' : 'none';
});

// 完成设置
function finishSetup() {
    // 验证所有步骤
    if (!validateCurrentStep(1) || !validateCurrentStep(2)) {
        return;
    }

    // 显示加载提示
    document.getElementById('loadingOverlay').style.display = 'flex';

    // 收集配置数据
    const config = {
        serverName: document.getElementById('serverName').value,
        dataDir: document.getElementById('dataDir').value,
        port: parseInt(document.getElementById('port').value),
        maxFileSize: parseInt(document.getElementById('maxFileSize').value),
        adminUsername: document.getElementById('adminUsername').value,
        adminPassword: document.getElementById('adminPassword').value,
        enableTLS: document.getElementById('enableTLS').checked,
        enableAuth: document.getElementById('enableAuth').checked
    };

    // 发送配置到服务器
    fetch('/api/setup/initialize', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(config)
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            // 显示成功提示
            document.getElementById('loadingOverlay').style.display = 'none';
            document.getElementById('successModal').style.display = 'flex';
        } else {
            throw new Error(data.message || '配置保存失败');
        }
    })
    .catch(error => {
        alert('配置保存失败: ' + error.message);
        document.getElementById('loadingOverlay').style.display = 'none';
    });
}

// 密码确认监听
document.getElementById('confirmPassword').addEventListener('input', function() {
    const password = document.getElementById('adminPassword').value;
    const passwordError = document.getElementById('passwordError');

    if (this.value && password !== this.value) {
        passwordError.style.display = 'block';
    } else {
        passwordError.style.display = 'none';
    }
});

// =================== 管理后台 JavaScript ===================

// 页面导航
document.addEventListener('DOMContentLoaded', function() {
    // 绑定导航点击事件
    const navItems = document.querySelectorAll('.nav-item');
    navItems.forEach(item => {
        item.addEventListener('click', function() {
            // 移除所有活动状态
            navItems.forEach(nav => nav.classList.remove('active'));
            pages.forEach(page => page.classList.remove('active'));

            // 设置当前活动状态
            this.classList.add('active');
            const pageId = this.getAttribute('data-page') + '-page';
            document.getElementById(pageId).classList.add('active');

            // 更新页面标题
            const pageTitle = this.querySelector('span').textContent;
            document.getElementById('pageTitle').textContent = pageTitle;
        });
    });
});

// 侧边栏折叠
let sidebarCollapsed = false;
const menuToggle = document.getElementById('menuToggle');
if (menuToggle) {
    menuToggle.addEventListener('click', function() {
    const sidebar = document.querySelector('.sidebar');
    sidebarCollapsed = !sidebarCollapsed;

    if (sidebarCollapsed) {
        sidebar.classList.add('collapsed');
    } else {
        sidebar.classList.remove('collapsed');
    }
    });
}

// 文件搜索
const fileSearch = document.getElementById('fileSearch');
if (fileSearch) {
    fileSearch.addEventListener('input', function() {
    const searchTerm = this.value.toLowerCase();
    const fileItems = document.querySelectorAll('.file-item');

    fileItems.forEach(item => {
        const fileName = item.querySelector('h3').textContent.toLowerCase();
        if (fileName.includes(searchTerm)) {
            item.style.display = 'block';
        } else {
            item.style.display = 'none';
        }
    });
    });
}

// 文件过滤
const filterButtons = document.querySelectorAll('.filter-buttons button');
filterButtons.forEach(button => {
    button.addEventListener('click', function() {
        // 更新按钮状态
        filterButtons.forEach(btn => btn.classList.remove('active'));
        this.classList.add('active');

        // 这里可以添加实际的过滤逻辑
        const filter = this.getAttribute('data-filter');
        filterFiles(filter);
    });
});

// 过滤文件
function filterFiles(filter) {
    const fileItems = document.querySelectorAll('.file-item');

    switch(filter) {
        case 'all':
            fileItems.forEach(item => item.style.display = 'block');
            break;
        case 'recent':
            // 显示最近24小时的文件
            fileItems.forEach(item => {
                const date = new Date(item.querySelector('p').textContent.split(' • ')[1]);
                const now = new Date();
                const diffHours = (now - date) / (1000 * 60 * 60);
                if (diffHours < 24) {
                    item.style.display = 'block';
                } else {
                    item.style.display = 'none';
                }
            });
            break;
        case 'large':
            // 显示大于10MB的文件
            fileItems.forEach(item => {
                const sizeText = item.querySelector('.file-stats span:first-child').textContent;
                const sizeMB = parseFloat(sizeText.match(/[\d.]+/)[0]);
                if (sizeMB > 10) {
                    item.style.display = 'block';
                } else {
                    item.style.display = 'none';
                }
            });
            break;
        case 'popular':
            // 显示下载量大于100的文件
            fileItems.forEach(item => {
                const downloads = parseInt(item.querySelector('.file-stats span:last-child').textContent.match(/[\d,]+/)[0].replace(',', ''));
                if (downloads > 100) {
                    item.style.display = 'block';
                } else {
                    item.style.display = 'none';
                }
            });
            break;
    }
}

// 上传文件功能
function uploadFile() {
    document.getElementById('uploadModal').classList.add('active');
    document.getElementById('uploadArea').style.display = 'block';
    document.getElementById('uploadProgress').style.display = 'none';
}

function closeModal(modalId) {
    document.getElementById(modalId).classList.remove('active');
}

// 拖拽上传
const uploadArea = document.getElementById('uploadArea');
const fileInput = document.getElementById('fileInput');

uploadArea.addEventListener('click', () => fileInput.click());

uploadArea.addEventListener('dragover', (e) => {
    e.preventDefault();
    uploadArea.style.background = '#e3f2fd';
});

uploadArea.addEventListener('dragleave', () => {
    uploadArea.style.background = '';
});

uploadArea.addEventListener('drop', (e) => {
    e.preventDefault();
    uploadArea.style.background = '';
    const files = e.dataTransfer.files;
    handleFiles(files);
});

fileInput.addEventListener('change', (e) => {
    handleFiles(e.target.files);
});

// 处理文件上传
function handleFiles(files) {
    if (files.length === 0) return;

    uploadArea.style.display = 'none';
    document.getElementById('uploadProgress').style.display = 'block';

    // 模拟上传进度
    let progress = 0;
    const progressBar = document.querySelector('.progress-fill');
    const progressText = document.querySelector('.progress-text');

    const interval = setInterval(() => {
        progress += 10;
        progressBar.style.width = progress + '%';
        progressText.textContent = `上传中... ${progress}%`;

        if (progress >= 100) {
            clearInterval(interval);
            setTimeout(() => {
                closeModal('uploadModal');
                alert('文件上传成功！');
                // 这里可以刷新文件列表
            }, 500);
        }
    }, 200);
}

// 下载文件
function downloadFile(filename) {
    window.open(`/api/download/${filename}`, '_blank');
}

// 分享文件
function shareFile(filename) {
    // 这里可以实现分享功能
    alert(`分享文件: ${filename}`);
}

// 删除文件
function deleteFile(filename) {
    if (confirm(`确定要删除文件 "${filename}" 吗？`)) {
        // 发送删除请求
        fetch(`/api/files/${filename}`, {
            method: 'DELETE'
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('文件删除成功！');
                // 这里可以刷新文件列表
            } else {
                alert('删除失败: ' + data.message);
            }
        })
        .catch(error => {
            alert('删除失败: ' + error.message);
        });
    }
}

// 添加用户
function addUser() {
    alert('添加用户功能');
}

// 退出登录
function logout() {
    if (confirm('确定要退出登录吗？')) {
        window.location.href = '/logout';
    }
}

// 更新系统状态
function updateSystemStatus() {
    // 模拟获取系统状态
    fetch('/api/status')
    .then(response => response.json())
    .then(data => {
        document.getElementById('memoryUsage').textContent = data.memory + '%';
        document.getElementById('diskUsage').textContent = data.disk + '%';
    })
    .catch(error => {
        console.error('获取系统状态失败:', error);
    });
}

// 定期更新系统状态
setInterval(updateSystemStatus, 30000); // 每30秒更新一次

// 初始化图表（如果有图表库）
function initCharts() {
    // 这里可以初始化图表，例如使用 Chart.js
    // 示例代码：
    /*
    const ctx = document.getElementById('storageChart').getContext('2d');
    new Chart(ctx, {
        type: 'line',
        data: {
            labels: ['1月', '2月', '3月', '4月', '5月', '6月'],
            datasets: [{
                label: '存储使用量',
                data: [10, 15, 20, 25, 30, 35],
                borderColor: '#3498db',
                backgroundColor: 'rgba(52, 152, 219, 0.1)',
                tension: 0.4
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        callback: function(value) {
                            return value + ' GB';
                        }
                    }
                }
            }
        }
    });
    */
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    initCharts();
    updateSystemStatus();
});

// WebSocket 连接（用于实时更新）
function initWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const ws = new WebSocket(`${protocol}//${window.location.host}/ws`);

    ws.onmessage = function(event) {
        const data = JSON.parse(event.data);
        handleWebSocketMessage(data);
    };

    ws.onclose = function() {
        // 10秒后重连
        setTimeout(initWebSocket, 10000);
    };
}

// 处理 WebSocket 消息
function handleWebSocketMessage(data) {
    switch(data.type) {
        case 'file_uploaded':
            // 处理文件上传事件
            break;
        case 'user_activity':
            // 处理用户活动事件
            break;
        case 'system_alert':
            // 处理系统警报
            break;
    }
}

// 启动 WebSocket
initWebSocket();