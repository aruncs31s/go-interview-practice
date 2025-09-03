package handlers

import (
	"encoding/json"
	"net/http"

	"web-ui/internal/models"
	"web-ui/internal/services"
	"web-ui/internal/utils"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// SignUpHandler handles user registration
func (h *AuthHandler) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// Parse JSON request
	var signupData models.UserSignup
	if err := json.NewDecoder(r.Body).Decode(&signupData); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON format"})
		return
	}

	// Create user
	response, err := h.authService.SignUp(&signupData)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to encode response"})
		return
	}
}

// LoginHandler handles user authentication
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// Parse JSON request
	var credentials models.UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON format"})
		return
	}

	// Authenticate user
	response, err := h.authService.Login(&credentials)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to encode response"})
		return
	}
}

// MeHandler returns current user info from token
func (h *AuthHandler) MeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	token := utils.ExtractTokenFromHeader(authHeader)
	if token == "" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	// Validate token and get user
	user, err := h.authService.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// LoginPageHandler serves the login HTML page
func (h *AuthHandler) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login - Go Interview Practice</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.0/font/bootstrap-icons.css">
    <style>
        /* Theme Variables */
        :root {
            --bg-gradient-start: #f8faff;
            --bg-gradient-end: #f0f4ff;
            --text-color: #2d3748;
            --text-muted: #718096;
            --card-bg: rgba(255, 255, 255, 0.9);
            --input-bg: rgba(255, 255, 255, 0.8);
        }
        
        [data-theme="dark"] {
            --bg-gradient-start: #0d1117;
            --bg-gradient-end: #1a1f29;
            --text-color: #e6edf3;
            --text-muted: #7d8590;
            --card-bg: rgba(22, 27, 34, 0.9);
            --input-bg: rgba(22, 27, 34, 0.8);
        }
        
        body {
            background: linear-gradient(135deg, var(--bg-gradient-start) 0%, var(--bg-gradient-end) 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            color: var(--text-color);
            transition: all 0.3s ease;
        }
        
        .auth-container {
            max-width: 420px;
            margin: 0 auto;
            padding: 2rem;
        }
        
        .auth-card {
            background: rgba(255, 255, 255, 0.9);
            backdrop-filter: blur(20px);
            border-radius: 20px;
            border: 1px solid rgba(255, 255, 255, 0.3);
            box-shadow: 0 20px 40px rgba(102, 126, 234, 0.1);
            padding: 2.5rem;
            position: relative;
            overflow: hidden;
        }
        
        .auth-card::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            height: 4px;
            background: linear-gradient(90deg, #667eea, #764ba2);
        }
        
        .auth-header {
            text-align: center;
            margin-bottom: 2rem;
        }
        
        .auth-title {
            color: #2d3748;
            font-weight: 700;
            font-size: 1.75rem;
            margin-bottom: 0.5rem;
        }
        
        .auth-subtitle {
            color: #718096;
            font-size: 0.95rem;
        }
        
        .form-group {
            margin-bottom: 1.5rem;
        }
        
        .form-label {
            color: #4a5568;
            font-weight: 600;
            margin-bottom: 0.5rem;
            font-size: 0.9rem;
        }
        
        .form-control {
            border: 2px solid #e2e8f0;
            border-radius: 12px;
            padding: 0.75rem 1rem;
            font-size: 1rem;
            transition: all 0.3s ease;
            background: rgba(255, 255, 255, 0.8);
        }
        
        .form-control:focus {
            border-color: #667eea;
            box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
            background: rgba(255, 255, 255, 0.95);
        }
        
        .form-control.is-invalid {
            border-color: #e53e3e;
            background: rgba(254, 226, 226, 0.5);
        }
        
        .auth-btn {
            width: 100%;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            border: none;
            border-radius: 12px;
            padding: 0.875rem 1.5rem;
            font-weight: 600;
            font-size: 1rem;
            color: white;
            transition: all 0.3s ease;
            box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3);
        }
        
        .auth-btn:hover {
            background: linear-gradient(135deg, #764ba2 0%, #667eea 100%);
            transform: translateY(-2px);
            box-shadow: 0 6px 20px rgba(102, 126, 234, 0.4);
        }
        
        .auth-btn:active {
            transform: translateY(0);
        }
        
        .auth-link {
            text-align: center;
            margin-top: 1.5rem;
            padding-top: 1.5rem;
            border-top: 1px solid rgba(226, 232, 240, 0.8);
        }
        
        .auth-link a {
            color: #667eea;
            text-decoration: none;
            font-weight: 500;
            transition: color 0.3s ease;
        }
        
        .auth-link a:hover {
            color: #764ba2;
            text-decoration: underline;
        }
        
        .message {
            margin-top: 1rem;
            padding: 0.75rem;
            border-radius: 8px;
            font-size: 0.9rem;
        }
        
        .message.error {
            background: rgba(254, 226, 226, 0.8);
            color: #c53030;
            border: 1px solid rgba(197, 48, 48, 0.3);
        }
        
        .message.success {
            background: rgba(198, 246, 213, 0.8);
            color: #276749;
            border: 1px solid rgba(39, 103, 73, 0.3);
        }
        
        .brand-link {
            display: inline-flex;
            align-items: center;
            text-decoration: none;
            color: #667eea;
            font-weight: 600;
            margin-bottom: 2rem;
            transition: color 0.3s ease;
        }
        
        .brand-link:hover {
            color: #764ba2;
        }
        
        .brand-link i {
            margin-right: 0.5rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="auth-container">
            <a href="/" class="brand-link">
                <i class="bi bi-arrow-left"></i>
                Back to Go Interview Practice
            </a>
            
            <div class="auth-card">
                <div class="auth-header">
                    <h1 class="auth-title">Welcome Back</h1>
                    <p class="auth-subtitle">Sign in to continue your coding journey</p>
                </div>
                
                <form id="loginForm">
                    <div class="form-group">
                        <label for="username" class="form-label">Username or Email</label>
                        <input type="text" class="form-control" id="username" name="username" required>
                    </div>
                    
                    <div class="form-group">
                        <label for="password" class="form-label">Password</label>
                        <input type="password" class="form-control" id="password" name="password" required>
                    </div>
                    
                    <button type="submit" class="auth-btn">
                        <i class="bi bi-box-arrow-in-right me-2"></i>
                        Sign In
                    </button>
                </form>
                
                <div id="message"></div>
                
                <div class="auth-link">
                    <span class="text-muted">Don't have an account? </span>
                    <a href="/signup">Create one here</a>
                </div>
            </div>
        </div>
    </div>

    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        // Initialize theme from localStorage
        const savedTheme = localStorage.getItem('theme') || 'light';
        document.documentElement.setAttribute('data-theme', savedTheme);
        
        document.getElementById('loginForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            // Clear previous error styling
            document.querySelectorAll('.form-control').forEach(input => {
                input.classList.remove('is-invalid');
            });
            
            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;
            
            try {
                const response = await fetch('/api/auth/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username, password }),
                });
                
                if (response.ok) {
                    const data = await response.json();
                    localStorage.setItem('token', data.token);
                    document.getElementById('message').innerHTML = '<div class="message success"><i class="bi bi-check-circle me-2"></i>Login successful! Redirecting...</div>';
                    setTimeout(() => { window.location.href = '/'; }, 1500);
                } else {
                    const errorText = await response.text();
                    let errorMessage = errorText;
                    
                    try {
                        const errorData = JSON.parse(errorText);
                        errorMessage = errorData.error || errorText;
                    } catch (e) {
                        // If not JSON, use the text as is
                    }
                    
                    // Highlight fields for login errors
                    document.getElementById('username').classList.add('is-invalid');
                    document.getElementById('password').classList.add('is-invalid');
                    
                    document.getElementById('message').innerHTML = '<div class="message error"><i class="bi bi-exclamation-circle me-2"></i>' + errorMessage + '</div>';
                }
            } catch (error) {
                document.getElementById('message').innerHTML = '<div class="message error"><i class="bi bi-wifi-off me-2"></i>Network error occurred</div>';
            }
        });
    </script>
</body>
</html>`))
}

// SignUpPageHandler serves the signup HTML page
func (h *AuthHandler) SignUpPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sign Up - Go Interview Practice</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.0/font/bootstrap-icons.css">
    <style>
        /* Theme Variables */
        :root {
            --bg-gradient-start: #f8faff;
            --bg-gradient-end: #f0f4ff;
            --text-color: #2d3748;
            --text-muted: #718096;
            --card-bg: rgba(255, 255, 255, 0.9);
            --input-bg: rgba(255, 255, 255, 0.8);
        }
        
        [data-theme="dark"] {
            --bg-gradient-start: #0d1117;
            --bg-gradient-end: #1a1f29;
            --text-color: #e6edf3;
            --text-muted: #7d8590;
            --card-bg: rgba(22, 27, 34, 0.9);
            --input-bg: rgba(22, 27, 34, 0.8);
        }
        
        body {
            background: linear-gradient(135deg, var(--bg-gradient-start) 0%, var(--bg-gradient-end) 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            color: var(--text-color);
            transition: all 0.3s ease;
        }
        
        .auth-container {
            max-width: 420px;
            margin: 0 auto;
            padding: 2rem;
        }
        
        .auth-card {
            background: rgba(255, 255, 255, 0.9);
            backdrop-filter: blur(20px);
            border-radius: 20px;
            border: 1px solid rgba(255, 255, 255, 0.3);
            box-shadow: 0 20px 40px rgba(102, 126, 234, 0.1);
            padding: 2.5rem;
            position: relative;
            overflow: hidden;
        }
        
        .auth-card::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            height: 4px;
            background: linear-gradient(90deg, #667eea, #764ba2);
        }
        
        .auth-header {
            text-align: center;
            margin-bottom: 2rem;
        }
        
        .auth-title {
            color: #2d3748;
            font-weight: 700;
            font-size: 1.75rem;
            margin-bottom: 0.5rem;
        }
        
        .auth-subtitle {
            color: #718096;
            font-size: 0.95rem;
        }
        
        .form-group {
            margin-bottom: 1.5rem;
        }
        
        .form-label {
            color: #4a5568;
            font-weight: 600;
            margin-bottom: 0.5rem;
            font-size: 0.9rem;
        }
        
        .form-control {
            border: 2px solid #e2e8f0;
            border-radius: 12px;
            padding: 0.75rem 1rem;
            font-size: 1rem;
            transition: all 0.3s ease;
            background: rgba(255, 255, 255, 0.8);
        }
        
        .form-control:focus {
            border-color: #667eea;
            box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
            background: rgba(255, 255, 255, 0.95);
        }
        
        .form-control.is-invalid {
            border-color: #e53e3e;
            background: rgba(254, 226, 226, 0.5);
        }
        
        .form-control.is-valid {
            border-color: #38a169;
            background: rgba(198, 246, 213, 0.5);
        }
        
        .auth-btn {
            width: 100%;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            border: none;
            border-radius: 12px;
            padding: 0.875rem 1.5rem;
            font-weight: 600;
            font-size: 1rem;
            color: white;
            transition: all 0.3s ease;
            box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3);
        }
        
        .auth-btn:hover {
            background: linear-gradient(135deg, #764ba2 0%, #667eea 100%);
            transform: translateY(-2px);
            box-shadow: 0 6px 20px rgba(102, 126, 234, 0.4);
        }
        
        .auth-btn:active {
            transform: translateY(0);
        }
        
        .auth-link {
            text-align: center;
            margin-top: 1.5rem;
            padding-top: 1.5rem;
            border-top: 1px solid rgba(226, 232, 240, 0.8);
        }
        
        .auth-link a {
            color: #667eea;
            text-decoration: none;
            font-weight: 500;
            transition: color 0.3s ease;
        }
        
        .auth-link a:hover {
            color: #764ba2;
            text-decoration: underline;
        }
        
        .password-requirements {
            background: rgba(240, 244, 255, 0.8);
            border: 1px solid rgba(102, 126, 234, 0.2);
            border-radius: 8px;
            padding: 0.75rem;
            margin-top: 0.5rem;
            font-size: 0.85rem;
        }
        
        .requirement {
            display: flex;
            align-items: center;
            margin: 0.25rem 0;
            color: #718096;
            transition: color 0.3s ease;
        }
        
        .requirement.met {
            color: #38a169;
        }
        
        .requirement i {
            width: 16px;
            margin-right: 0.5rem;
            font-size: 0.75rem;
        }
        
        .message {
            margin-top: 1rem;
            padding: 0.75rem;
            border-radius: 8px;
            font-size: 0.9rem;
        }
        
        .message.error {
            background: rgba(254, 226, 226, 0.8);
            color: #c53030;
            border: 1px solid rgba(197, 48, 48, 0.3);
        }
        
        .message.success {
            background: rgba(198, 246, 213, 0.8);
            color: #276749;
            border: 1px solid rgba(39, 103, 73, 0.3);
        }
        
        .brand-link {
            display: inline-flex;
            align-items: center;
            text-decoration: none;
            color: #667eea;
            font-weight: 600;
            margin-bottom: 2rem;
            transition: color 0.3s ease;
        }
        
        .brand-link:hover {
            color: #764ba2;
        }
        
        .brand-link i {
            margin-right: 0.5rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="auth-container">
            <a href="/" class="brand-link">
                <i class="bi bi-arrow-left"></i>
                Back to Go Interview Practice
            </a>
            
            <div class="auth-card">
                <div class="auth-header">
                    <h1 class="auth-title">Create Account</h1>
                    <p class="auth-subtitle">Join us and start your coding journey</p>
                </div>
                
                <form id="signupForm">
                    <div class="form-group">
                        <label for="username" class="form-label">Username</label>
                        <input type="text" class="form-control" id="username" name="username" required>
                    </div>
                    
                    <div class="form-group">
                        <label for="email" class="form-label">Email Address</label>
                        <input type="email" class="form-control" id="email" name="email" required>
                    </div>
                    
                    <div class="form-group">
                        <label for="password" class="form-label">Password</label>
                        <input type="password" class="form-control" id="password" name="password" required>
                        <div class="password-requirements">
                            <div class="requirement" id="req-length">
                                <i class="bi bi-circle"></i>
                                At least 8 characters
                            </div>
                            <div class="requirement" id="req-upper">
                                <i class="bi bi-circle"></i>
                                One uppercase letter
                            </div>
                            <div class="requirement" id="req-lower">
                                <i class="bi bi-circle"></i>
                                One lowercase letter
                            </div>
                            <div class="requirement" id="req-digit">
                                <i class="bi bi-circle"></i>
                                One number
                            </div>
                        </div>
                    </div>
                    
                    <button type="submit" class="auth-btn">
                        <i class="bi bi-person-plus me-2"></i>
                        Create Account
                    </button>
                </form>
                
                <div id="message"></div>
                
                <div class="auth-link">
                    <span class="text-muted">Already have an account? </span>
                    <a href="/login">Sign in here</a>
                </div>
            </div>
        </div>
    </div>

    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        // Initialize theme from localStorage
        const savedTheme = localStorage.getItem('theme') || 'light';
        document.documentElement.setAttribute('data-theme', savedTheme);
        
        // Password validation in real-time
        document.getElementById('password').addEventListener('input', function() {
            const password = this.value;
            
            // Check each requirement
            const requirements = {
                'req-length': password.length >= 8,
                'req-upper': /[A-Z]/.test(password),
                'req-lower': /[a-z]/.test(password),
                'req-digit': /\d/.test(password)
            };
            
            // Update UI for each requirement
            Object.entries(requirements).forEach(([id, met]) => {
                const element = document.getElementById(id);
                const icon = element.querySelector('i');
                
                if (met) {
                    element.classList.add('met');
                    icon.className = 'bi bi-check-circle-fill';
                } else {
                    element.classList.remove('met');
                    icon.className = 'bi bi-circle';
                }
            });
            
            // Update password field styling
            const allMet = Object.values(requirements).every(met => met);
            if (password && allMet) {
                this.classList.remove('is-invalid');
                this.classList.add('is-valid');
            } else if (password) {
                this.classList.remove('is-valid');
                this.classList.add('is-invalid');
            } else {
                this.classList.remove('is-valid', 'is-invalid');
            }
        });

        document.getElementById('signupForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            // Clear previous error styling
            document.querySelectorAll('.form-control').forEach(input => {
                if (!input.classList.contains('is-valid')) {
                    input.classList.remove('is-invalid');
                }
            });
            
            const username = document.getElementById('username').value;
            const email = document.getElementById('email').value;
            const password = document.getElementById('password').value;
            
            try {
                const response = await fetch('/api/auth/signup', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username, email, password }),
                });
                
                if (response.ok) {
                    const data = await response.json();
                    localStorage.setItem('token', data.token);
                    document.getElementById('message').innerHTML = '<div class="message success"><i class="bi bi-check-circle me-2"></i>Account created successfully! Redirecting...</div>';
                    setTimeout(() => { window.location.href = '/'; }, 2000);
                } else {
                    // Handle error response
                    const errorText = await response.text();
                    let errorMessage = errorText;
                    
                    // Try to parse as JSON for better error messages
                    try {
                        const errorData = JSON.parse(errorText);
                        errorMessage = errorData.error || errorText;
                    } catch (e) {
                        // If not JSON, use the text as is
                    }
                    
                    // Highlight relevant fields based on error
                    if (errorMessage.includes('username')) {
                        document.getElementById('username').classList.add('is-invalid');
                    }
                    if (errorMessage.includes('email')) {
                        document.getElementById('email').classList.add('is-invalid');
                    }
                    if (errorMessage.includes('password')) {
                        document.getElementById('password').classList.add('is-invalid');
                    }
                    
                    document.getElementById('message').innerHTML = '<div class="message error"><i class="bi bi-exclamation-circle me-2"></i>' + errorMessage + '</div>';
                }
            } catch (error) {
                document.getElementById('message').innerHTML = '<div class="message error"><i class="bi bi-wifi-off me-2"></i>Network error occurred</div>';
            }
        });
    </script>
</body>
</html>`))
}
