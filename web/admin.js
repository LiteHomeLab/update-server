// ==================== 管理后台 JavaScript ====================

// 状态管理
let currentProgram = null;  // 当前选中的程序
let programList = [];       // 程序列表
let currentUploadToken = null;  // 当前程序的 upload token

// ==================== 初始化 ====================

document.addEventListener('DOMContentLoaded', function() {
    loadPrograms();
    setupEventListeners();
});

function setupEventListeners() {
    // 创建程序按钮
    document.getElementById('createProgramBtn').addEventListener('click', function() {
        document.getElementById('createProgramModal').classList.add('active');
    });

    // 创建程序表单
    document.getElementById('createProgramForm').addEventListener('submit', handleCreateProgram);

    // 上传版本表单
    document.getElementById('uploadVersionForm').addEventListener('submit', handleUploadVersion);

    // 上传区域点击
    const uploadArea = document.getElementById('uploadArea');
    const fileInput = document.getElementById('versionFile');
    uploadArea.addEventListener('click', () => fileInput.click());
    fileInput.addEventListener('change', handleFileSelect);

    // 拖拽上传
    uploadArea.addEventListener('dragover', (e) => {
        e.preventDefault();
        uploadArea.style.borderColor = '#667eea';
        uploadArea.style.background = '#f0f4ff';
    });
    uploadArea.addEventListener('dragleave', () => {
        uploadArea.style.borderColor = '#dee2e6';
        uploadArea.style.background = '#f8f9fa';
    });
    uploadArea.addEventListener('drop', (e) => {
        e.preventDefault();
        uploadArea.style.borderColor = '#dee2e6';
        uploadArea.style.background = '#f8f9fa';
        if (e.dataTransfer.files.length > 0) {
            fileInput.files = e.dataTransfer.files;
            handleFileSelect({ target: fileInput });
        }
    });
}

// ==================== 程序列表管理 ====================

async function loadPrograms() {
    try {
        const response = await fetch('/api/admin/programs');
        if (!response.ok) {
            throw new Error('加载程序列表失败');
        }
        programList = await response.json();
        renderProgramList();

        // 如果有程序且当前未选中，选中第一个
        if (programList.length > 0 && !currentProgram) {
            selectProgram(programList[0].programId);
        } else if (programList.length === 0) {
            showEmptyState();
        }
    } catch (error) {
        showToast('加载程序列表失败: ' + error.message, 'error');
    }
}

function renderProgramList() {
    const container = document.getElementById('programList');

    if (programList.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <i class="fas fa-folder-open"></i>
                <p>暂无程序</p>
            </div>
        `;
        return;
    }

    // 清空容器
    container.innerHTML = '';

    // 使用安全的 DOM 方法创建元素
    programList.forEach(program => {
        const card = document.createElement('div');
        card.className = 'program-card';
        if (currentProgram === program.programId) {
            card.classList.add('active');
        }
        card.onclick = () => selectProgram(program.programId);

        const nameDiv = document.createElement('div');
        nameDiv.className = 'program-card-name';
        nameDiv.textContent = program.name;

        const idDiv = document.createElement('div');
        idDiv.className = 'program-card-id';
        idDiv.textContent = program.programId;

        card.appendChild(nameDiv);
        card.appendChild(idDiv);
        container.appendChild(card);
    });
}

async function selectProgram(programId) {
    currentProgram = programId;

    // 更新列表高亮
    renderProgramList();

    // 显示内容区域
    document.getElementById('emptyState').style.display = 'none';
    document.getElementById('pageContent').style.display = 'block';

    try {
        const response = await fetch(`/api/admin/programs/${programId}`);
        if (!response.ok) {
            throw new Error('加载程序详情失败');
        }
        const data = await response.json();

        // 保存 upload token 用于后续上传
        currentUploadToken = data.uploadToken;

        // 更新各模块
        updateTokensModule(data);
        loadVersions(programId);
        updateCommandExamples(data);

        // 更新页面标题
        document.getElementById('pageTitle').textContent = data.program.name;
    } catch (error) {
        showToast('加载程序详情失败: ' + error.message, 'error');
    }
}

function showEmptyState() {
    document.getElementById('emptyState').style.display = 'flex';
    document.getElementById('pageContent').style.display = 'none';
    document.getElementById('pageTitle').textContent = '程序管理';
}

// ==================== 创建程序 ====================

async function handleCreateProgram(e) {
    e.preventDefault();

    const form = e.target;
    const name = document.getElementById('programName').value.trim();

    if (!name) {
        showToast('请输入程序名称', 'error');
        return;
    }

    try {
        const response = await fetch('/api/admin/programs', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name: name })
        });

        if (!response.ok) {
            throw new Error('创建程序失败');
        }

        const result = await response.json();

        // 添加到列表
        programList.push(result.program);

        // 选中新程序
        selectProgram(result.program.programId);

        // 关闭模态框
        closeModal('createProgramModal');
        form.reset();

        showToast('程序创建成功', 'success');
    } catch (error) {
        showToast('创建程序失败: ' + error.message, 'error');
    }
}

// ==================== Token & 密钥管理 ====================

function updateTokensModule(data) {
    document.getElementById('uploadToken').textContent = data.uploadToken || 'loading...';
    document.getElementById('downloadToken').textContent = data.downloadToken || 'loading...';
    document.getElementById('encryptionKey').textContent = data.encryptionKey || 'loading...';
}

async function copyToken(type) {
    let text = '';
    switch (type) {
        case 'upload':
            text = document.getElementById('uploadToken').textContent;
            break;
        case 'download':
            text = document.getElementById('downloadToken').textContent;
            break;
        case 'encryption':
            text = document.getElementById('encryptionKey').textContent;
            break;
    }

    if (!text || text === 'loading...') {
        showToast('内容未加载', 'error');
        return;
    }

    try {
        await navigator.clipboard.writeText(text);
        showToast('已复制到剪贴板', 'success');
    } catch (err) {
        // 降级方案
        const textarea = document.createElement('textarea');
        textarea.value = text;
        textarea.style.position = 'fixed';
        textarea.style.opacity = '0';
        document.body.appendChild(textarea);
        textarea.select();
        try {
            document.execCommand('copy');
            showToast('已复制到剪贴板', 'success');
        } catch (e) {
            showToast('复制失败', 'error');
        }
        document.body.removeChild(textarea);
    }
}

async function regenerateToken(type) {
    if (!currentProgram) return;

    if (!confirm(`确定要重新生成 ${type} token 吗？`)) {
        return;
    }

    try {
        const response = await fetch(`/api/admin/programs/${currentProgram}/tokens/regenerate?type=${type}`, {
            method: 'POST'
        });

        if (!response.ok) {
            throw new Error('重新生成 token 失败');
        }

        const data = await response.json();

        // 更新显示
        if (type === 'upload') {
            document.getElementById('uploadToken').textContent = data.token;
            currentUploadToken = data.token;
        } else {
            document.getElementById('downloadToken').textContent = data.token;
        }

        showToast('Token 已重新生成', 'success');
    } catch (error) {
        showToast('重新生成 token 失败: ' + error.message, 'error');
    }
}

async function regenerateEncryptionKey() {
    if (!currentProgram) return;

    if (!confirm('确定要重新生成加密密钥吗？重新生成后旧版本将无法验证！')) {
        return;
    }

    try {
        const response = await fetch(`/api/admin/programs/${currentProgram}/encryption/regenerate`, {
            method: 'POST'
        });

        if (!response.ok) {
            throw new Error('重新生成密钥失败');
        }

        const data = await response.json();
        document.getElementById('encryptionKey').textContent = data.encryptionKey;

        showToast('加密密钥已重新生成', 'success');
    } catch (error) {
        showToast('重新生成密钥失败: ' + error.message, 'error');
    }
}

// ==================== 版本管理 ====================

async function loadVersions(programId) {
    const tbody = document.getElementById('versionsTableBody');
    tbody.innerHTML = '<tr><td colspan="6" class="loading-row"><i class="fas fa-spinner fa-spin"></i> 加载中...</td></tr>';

    try {
        const response = await fetch(`/api/admin/programs/${programId}/versions`);
        if (!response.ok) {
            throw new Error('加载版本列表失败');
        }
        const versions = await response.json();
        renderVersions(versions);
    } catch (error) {
        tbody.innerHTML = '<tr><td colspan="6" class="loading-row">加载失败</td></tr>';
        showToast('加载版本列表失败: ' + error.message, 'error');
    }
}

function renderVersions(versions) {
    const tbody = document.getElementById('versionsTableBody');

    if (!versions || versions.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="loading-row">暂无版本</td></tr>';
        return;
    }

    // 清空并使用安全的方式渲染
    tbody.innerHTML = '';

    versions.forEach((v, index) => {
        const tr = document.createElement('tr');
        if (index === 0) {
            tr.classList.add('latest');
        }

        // 版本号
        const tdVersion = document.createElement('td');
        tdVersion.textContent = v.version;

        // 通道
        const tdChannel = document.createElement('td');
        const badge = document.createElement('span');
        badge.className = `badge ${v.channel}`;
        badge.textContent = v.channel === 'stable' ? '稳定版' : '测试版';
        tdChannel.appendChild(badge);

        // 大小
        const tdSize = document.createElement('td');
        tdSize.textContent = formatFileSize(v.filesize);

        // 发布时间
        const tdDate = document.createElement('td');
        tdDate.textContent = formatDate(v.publishDate);

        // 下载次数
        const tdDownloads = document.createElement('td');
        tdDownloads.textContent = (v.downloadCount || 0).toString();

        // 操作
        const tdActions = document.createElement('td');
        const deleteBtn = document.createElement('button');
        deleteBtn.className = 'btn btn-sm btn-danger';
        deleteBtn.innerHTML = '<i class="fas fa-trash"></i>';
        deleteBtn.onclick = () => deleteVersion(v.version);
        tdActions.appendChild(deleteBtn);

        tr.appendChild(tdVersion);
        tr.appendChild(tdChannel);
        tr.appendChild(tdSize);
        tr.appendChild(tdDate);
        tr.appendChild(tdDownloads);
        tr.appendChild(tdActions);
        tbody.appendChild(tr);
    });
}

async function deleteVersion(version) {
    if (!currentProgram) return;

    if (!confirm(`确定要删除版本 ${version} 吗？`)) {
        return;
    }

    try {
        const response = await fetch(`/api/admin/programs/${currentProgram}/versions/${version}`, {
            method: 'DELETE'
        });

        if (!response.ok) {
            throw new Error('删除版本失败');
        }

        showToast('版本已删除', 'success');
        loadVersions(currentProgram);
    } catch (error) {
        showToast('删除版本失败: ' + error.message, 'error');
    }
}

function showUploadModal() {
    document.getElementById('uploadVersionModal').classList.add('active');
}

function handleFileSelect(e) {
    const file = e.target.files[0];
    if (file) {
        const uploadArea = document.getElementById('uploadArea');
        uploadArea.innerHTML = '';
        const icon = document.createElement('i');
        icon.className = 'fas fa-file-archive';
        const p1 = document.createElement('p');
        p1.innerHTML = `<strong>${escapeHtml(file.name)}</strong>`;
        const p2 = document.createElement('p');
        p2.textContent = formatFileSize(file.size);
        uploadArea.appendChild(icon);
        uploadArea.appendChild(p1);
        uploadArea.appendChild(p2);
    }
}

async function handleUploadVersion(e) {
    e.preventDefault();

    if (!currentProgram || !currentUploadToken) {
        showToast('请先选择程序', 'error');
        return;
    }

    const fileInput = document.getElementById('versionFile');
    const file = fileInput.files[0];

    if (!file) {
        showToast('请选择文件', 'error');
        return;
    }

    const version = document.getElementById('versionNumber').value.trim();
    const channel = document.getElementById('versionChannel').value;
    const notes = document.getElementById('versionNotes').value.trim();
    const mandatory = document.getElementById('mandatoryUpdate').checked;

    if (!version) {
        showToast('请输入版本号', 'error');
        return;
    }

    const formData = new FormData();
    formData.append('file', file);
    formData.append('version', version);
    formData.append('channel', channel);
    formData.append('releaseNotes', notes);
    formData.append('mandatoryUpdate', mandatory);

    try {
        showToast('正在上传...', 'warning');

        const response = await fetch(`/api/programs/${currentProgram}/versions`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${currentUploadToken}`
            },
            body: formData
        });

        if (!response.ok) {
            throw new Error('上传版本失败');
        }

        showToast('版本上传成功', 'success');
        closeModal('uploadVersionModal');
        document.getElementById('uploadVersionForm').reset();
        loadVersions(currentProgram);
    } catch (error) {
        showToast('上传版本失败: ' + error.message, 'error');
    }
}

// ==================== 客户端工具 ====================

async function downloadUpdateClient() {
    if (!currentProgram) {
        showToast('请先选择程序', 'error');
        return;
    }

    try {
        const response = await fetch(`/api/admin/programs/${currentProgram}/client/update`);

        if (!response.ok) {
            throw new Error('下载客户端包失败');
        }

        // 获取文件名
        const disposition = response.headers.get('Content-Disposition');
        let filename = 'update-client.zip';
        if (disposition) {
            const match = disposition.match(/filename="(.+)"/);
            if (match) filename = match[1];
        }

        // 下载文件
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);

        showToast('下载已开始', 'success');
    } catch (error) {
        showToast('下载失败: ' + error.message, 'error');
    }
}

function updateCommandExamples(data) {
    const serverUrl = window.location.origin;
    const uploadToken = data.uploadToken;
    const programId = data.program.programId;

    const command = `# 上传新版本
./update-publisher upload \\
  --server=${serverUrl} \\
  --program-id=${programId} \\
  --token=${uploadToken} \\
  --file=./your-app.zip \\
  --version=1.0.0 \\
  --channel=stable \\
  --notes="修复了若干bug"`;

    document.getElementById('publisher-command').textContent = command;
}

async function copyCommand() {
    const command = document.getElementById('publisher-command').textContent;

    try {
        await navigator.clipboard.writeText(command);
        showToast('命令已复制到剪贴板', 'success');
    } catch (err) {
        // 降级方案
        const textarea = document.createElement('textarea');
        textarea.value = command;
        textarea.style.position = 'fixed';
        textarea.style.opacity = '0';
        document.body.appendChild(textarea);
        textarea.select();
        try {
            document.execCommand('copy');
            showToast('命令已复制到剪贴板', 'success');
        } catch (e) {
            showToast('复制失败', 'error');
        }
        document.body.removeChild(textarea);
    }
}

// ==================== 通用功能 ====================

function closeModal(modalId) {
    document.getElementById(modalId).classList.remove('active');
    // 重置表单
    const form = document.getElementById(modalId).querySelector('form');
    if (form) form.reset();

    // 重置上传区域
    if (modalId === 'uploadVersionModal') {
        const uploadArea = document.getElementById('uploadArea');
        uploadArea.innerHTML = '';
        const icon = document.createElement('i');
        icon.className = 'fas fa-cloud-upload-alt';
        const p = document.createElement('p');
        p.textContent = '拖拽文件到此处或点击选择文件';
        uploadArea.appendChild(icon);
        uploadArea.appendChild(p);
    }
}

function logout() {
    if (confirm('确定要退出登录吗？')) {
        window.location.href = '/admin/logout';
    }
}

// Toast 通知
function showToast(message, type = 'success') {
    const container = document.getElementById('toastContainer');
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;

    const icons = {
        success: 'fa-check-circle',
        error: 'fa-exclamation-circle',
        warning: 'fa-exclamation-triangle'
    };

    const icon = document.createElement('i');
    icon.className = `fas ${icons[type]} toast-icon`;

    const messageSpan = document.createElement('span');
    messageSpan.className = 'toast-message';
    messageSpan.textContent = message;

    toast.appendChild(icon);
    toast.appendChild(messageSpan);
    container.appendChild(toast);

    // 自动移除
    const timeout = type === 'error' ? 5000 : 2000;
    setTimeout(() => {
        toast.classList.add('removing');
        setTimeout(() => {
            if (toast.parentElement) {
                toast.parentElement.removeChild(toast);
            }
        }, 300);
    }, timeout);
}

// 工具函数
function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function formatFileSize(bytes) {
    if (!bytes) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

function formatDate(dateString) {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    });
}
