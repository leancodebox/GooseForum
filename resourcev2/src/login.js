// Vue 3 登录页面逻辑
import { createApp } from 'vue'
import './style.css'

// 确保DOM加载完成后再创建Vue实例
function initVueApp() {
    const app = createApp({
    data() {
        return {
            // 当前激活的标签页
            activeTab: 'login',
            
            // 登录表单数据
            loginForm: {
                username: '',
                password: '',
                remember: false
            },
            
            // 注册表单数据
            registerForm: {
                username: '',
                email: '',
                password: '',
                confirmPassword: '',
                agree: false
            },
            
            // 登录表单错误
            loginErrors: {
                username: '',
                password: ''
            },
            
            // 注册表单错误
            registerErrors: {
                username: '',
                email: '',
                password: '',
                confirmPassword: ''
            },
            
            // 加载状态
            loginLoading: false,
            registerLoading: false
        }
    },
    
    mounted() {
        // 从Go模板获取初始值
        const loginUsernameInput = document.querySelector('#login input[name="username"]');
        if (loginUsernameInput && loginUsernameInput.value) {
            this.loginForm.username = loginUsernameInput.value;
        }
        
        const registerUsernameInput = document.querySelector('#register input[name="username"]');
        if (registerUsernameInput && registerUsernameInput.value) {
            this.registerForm.username = registerUsernameInput.value;
        }
        
        const emailInput = document.querySelector('#register input[name="email"]');
        if (emailInput && emailInput.value) {
            this.registerForm.email = emailInput.value;
        }
    },
    
    methods: {
        // 切换标签页
        switchTab(tab) {
            this.activeTab = tab;
            // 清空错误信息
            this.loginErrors = { username: '', password: '' };
            this.registerErrors = { username: '', email: '', password: '', confirmPassword: '' };
        },
        
        // 获取CSRF令牌
        getCSRFToken() {
            const csrfMeta = document.querySelector('meta[name="csrf-token"]');
            return csrfMeta ? csrfMeta.getAttribute('content') : '';
        },
        // 验证登录表单
        validateLoginForm() {
            this.loginErrors = {
                username: '',
                password: ''
            };
            
            let isValid = true;
            
            if (!this.loginForm.username.trim()) {
                this.loginErrors.username = '请输入用户名或邮箱';
                isValid = false;
            }
            
            if (!this.loginForm.password.trim()) {
                this.loginErrors.password = '请输入密码';
                isValid = false;
            }
            
            return isValid;
        },
        
        // 验证注册表单
        validateRegisterForm() {
            this.registerErrors = {
                username: '',
                email: '',
                password: '',
                confirmPassword: ''
            };
            
            let isValid = true;
            
            if (!this.registerForm.username.trim()) {
                this.registerErrors.username = '请输入用户名';
                isValid = false;
            } else if (this.registerForm.username.length < 3) {
                this.registerErrors.username = '用户名至少3个字符';
                isValid = false;
            }
            
            if (!this.registerForm.email.trim()) {
                this.registerErrors.email = '请输入邮箱地址';
                isValid = false;
            } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(this.registerForm.email)) {
                this.registerErrors.email = '请输入有效的邮箱地址';
                isValid = false;
            }
            
            if (!this.registerForm.password.trim()) {
                this.registerErrors.password = '请输入密码';
                isValid = false;
            } else if (this.registerForm.password.length < 6) {
                this.registerErrors.password = '密码至少6个字符';
                isValid = false;
            }
            
            if (!this.registerForm.confirmPassword.trim()) {
                this.registerErrors.confirmPassword = '请确认密码';
                isValid = false;
            } else if (this.registerForm.password !== this.registerForm.confirmPassword) {
                this.registerErrors.confirmPassword = '两次输入的密码不一致';
                isValid = false;
            }
            
            return isValid;
        },
        
        // 处理登录
        async handleLogin() {
            if (!this.validateLoginForm()) {
                return;
            }
            
            this.loginLoading = true;
            
            try {
                const formData = new FormData();
                formData.append('username', this.loginForm.username);
                formData.append('password', this.loginForm.password);
                formData.append('remember', this.loginForm.remember ? '1' : '0');
                
                const csrfToken = this.getCSRFToken();
                if (csrfToken) {
                    formData.append('_token', csrfToken);
                }
                
                const response = await fetch('/login', {
                    method: 'POST',
                    body: formData,
                    headers: {
                        'X-Requested-With': 'XMLHttpRequest'
                    }
                });
                
                if (response.ok) {
                    const result = await response.json();
                    if (result.success) {
                        // 登录成功，重定向
                        window.location.href = result.redirect || '/dashboard';
                    } else {
                        // 显示错误信息
                        if (result.errors) {
                            this.loginErrors = { ...this.loginErrors, ...result.errors };
                        } else {
                            this.loginErrors.username = result.message || '登录失败';
                        }
                    }
                } else {
                    // HTTP错误
                    this.loginErrors.username = '网络错误，请稍后重试';
                }
            } catch (error) {
                console.error('登录错误:', error);
                this.loginErrors.username = '网络错误，请稍后重试';
            } finally {
                this.loginLoading = false;
            }
        },
        
        // 处理注册
        async handleRegister() {
            if (!this.validateRegisterForm()) {
                return;
            }
            
            if (!this.registerForm.agree) {
                alert('请同意用户协议和隐私政策');
                return;
            }
            
            this.registerLoading = true;
            
            try {
                const formData = new FormData();
                formData.append('username', this.registerForm.username);
                formData.append('email', this.registerForm.email);
                formData.append('password', this.registerForm.password);
                formData.append('confirm_password', this.registerForm.confirmPassword);
                
                const csrfToken = this.getCSRFToken();
                if (csrfToken) {
                    formData.append('_token', csrfToken);
                }
                
                const response = await fetch('/register', {
                    method: 'POST',
                    body: formData,
                    headers: {
                        'X-Requested-With': 'XMLHttpRequest'
                    }
                });
                
                if (response.ok) {
                    const result = await response.json();
                    if (result.success) {
                        // 注册成功
                        alert('注册成功！请查收邮箱验证邮件。');
                        // 切换到登录页面
                        this.switchTab('login');
                        // 清空注册表单
                        this.registerForm = {
                            username: '',
                            email: '',
                            password: '',
                            confirmPassword: '',
                            agree: false
                        };
                    } else {
                        // 显示错误信息
                        if (result.errors) {
                            this.registerErrors = { ...this.registerErrors, ...result.errors };
                        } else {
                            this.registerErrors.username = result.message || '注册失败';
                        }
                    }
                } else {
                    // HTTP错误
                    this.registerErrors.username = '网络错误，请稍后重试';
                }
            } catch (error) {
                console.error('注册错误:', error);
                this.registerErrors.username = '网络错误，请稍后重试';
            } finally {
                this.registerLoading = false;
            }
        }
    }});
    
    // 挂载Vue实例
    app.mount('#login');
    console.log('Vue应用已挂载到 #login');
}

// 确保DOM加载完成后再初始化Vue
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initVueApp);
} else {
    initVueApp();
}
