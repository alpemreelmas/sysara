# Sysara - Futuristic System Management Platform

![Sysara Logo](https://via.placeholder.com/400x100/667eea/ffffff?text=Sysara)

**Sysara** is a modern, web-based system management platform built with Go, designed to provide comprehensive server administration, monitoring, and configuration management through an intuitive web interface.

## ✨ Features

### 🔐 User Management
- **User Authentication**: Secure login/logout with session management
- **User CRUD Operations**: Create, read, update, and delete user accounts
- **Password Security**: BCrypt password hashing
- **Session Management**: Secure session handling with Gorilla Sessions

### 🌍 Environment Configuration
- **Multi-Environment Support**: Manage `.env`, `.env.production`, `.env.testing`, etc.
- **Web-Based Editor**: Edit environment files directly from the web interface
- **Automatic Backup**: Creates backups before saving changes
- **Syntax Validation**: Basic validation for environment file format

### 🔑 SSH Key Management
- **Key Storage**: Securely store and manage SSH public keys
- **Format Validation**: Validates SSH key formats (RSA, Ed25519, ECDSA)
- **Fingerprint Generation**: Automatic fingerprint generation
- **User Association**: Keys are associated with specific users

### 📊 System Monitoring
- **Real-Time Metrics**: Live CPU, Memory, Disk, and Network monitoring
- **Process Management**: View running processes with CPU and memory usage
- **System Information**: Display host information, uptime, and OS details
- **Auto-Refresh**: HTMX-powered automatic updates every 5 seconds

### 🎨 Modern UI/UX
- **Responsive Design**: Built with Tailwind CSS for mobile-first design
- **Interactive Elements**: HTMX for seamless user interactions
- **Real-Time Updates**: Dynamic content updates without page reloads
- **Accessibility**: WCAG-compliant interface design

## 🛠️ Technology Stack

- **Backend**: Go 1.21+ with Gin web framework
- **Database**: SQLite with GORM ORM
- **Frontend**: HTML templates + HTMX + Tailwind CSS
- **Authentication**: Gorilla Sessions with BCrypt
- **Monitoring**: gopsutil for system metrics
- **Deployment**: Systemd service with installer script

## 📋 Requirements

- Go 1.21 or higher
- SQLite3
- Linux/Unix system (for deployment)
- Modern web browser

## 🚀 Quick Start

### Development

1. **Clone the repository** (or use existing files):
```bash
git clone https://github.com/alpemreelmas/sysara.git
cd sysara
```

2. **Install dependencies**:
```bash
go mod tidy
```

3. **Build the application**:
```bash
# For Linux
go build .

# For Windows (disable CGO)
CGO_ENABLED=0 go build .
```

4. **Run the application**:
```bash
./sysara
```

5. **Access the application**:
Open your browser and navigate to `http://localhost:8080`

### Production Deployment (Linux)

1. **Make installer executable**:
```bash
chmod +x installer.sh
```

2. **Run the installer**:
```bash
sudo ./installer.sh
```

3. **Access the application**:
Navigate to `http://your-server-ip:8080`

## 📁 Project Structure

```
sysara/
├── cmd/                    # Command-line applications
├── config/                 # Configuration files
├── data/                   # Database and data files
├── internal/
│   ├── auth/              # Authentication logic
│   ├── handlers/          # HTTP handlers
│   ├── middleware/        # HTTP middleware
│   ├── models/            # Database models
│   └── services/          # Business logic services
├── logs/                  # Application logs
├── pkg/
│   └── utils/             # Utility functions
├── static/
│   ├── css/               # Stylesheets
│   ├── js/                # JavaScript files
│   └── images/            # Static images
├── templates/
│   ├── layouts/           # Layout templates
│   ├── pages/             # Page templates
│   └── partials/          # Partial templates
├── installer.sh           # Linux installation script
├── main.go               # Application entry point
└── README.md             # This file
```

## 🔧 Configuration

### Environment Variables

Create a `.env` file in the root directory:

```env
# Server Configuration
PORT=8080
GIN_MODE=release

# Database
DATABASE_PATH=data/sysara.db

# Security
SESSION_SECRET=your-secret-key-here

# Monitoring
REFRESH_INTERVAL=5000
```

### Default Configuration

The application will create default configurations on first run:
- SQLite database in `data/sysara.db`
- Log files in `logs/` directory
- Static files served from `static/` directory

## 🔐 Security Features

- **Password Hashing**: BCrypt with salt
- **Session Security**: Secure cookie settings
- **Input Validation**: Server-side validation for all inputs
- **CSRF Protection**: Built-in CSRF protection
- **SSH Key Validation**: Format validation for SSH keys

## 📖 API Documentation

### Authentication Endpoints

- `GET /login` - Display login page
- `POST /login` - Authenticate user
- `GET /register` - Display registration page
- `POST /register` - Create new user account
- `POST /logout` - Logout current user

### Protected Endpoints

- `GET /dashboard` - Main dashboard
- `GET /users` - List all users
- `GET /env` - Environment file management
- `GET /ssh` - SSH key management
- `GET /monitor` - System monitoring dashboard

### API Endpoints (HTMX)

- `GET /monitor/api/stats` - System statistics
- `GET /monitor/api/processes` - Running processes

## 🔄 Development

### Adding New Features

1. **Create Handler**: Add new handler in `internal/handlers/`
2. **Add Routes**: Update `main.go` with new routes
3. **Create Templates**: Add HTML templates in `templates/`
4. **Update Models**: Modify database models if needed

### Running in Development Mode

```bash
# Set development mode
export GIN_MODE=debug

# Run with auto-reload (requires air)
air

# Or run normally
go run main.go
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./internal/auth/
```

## 📦 Building for Production

### Linux
```bash
GOOS=linux GOARCH=amd64 go build -o sysara-linux .
```

### Windows
```bash
GOOS=windows GOARCH=amd64 go build -o sysara-windows.exe .
```

### macOS
```bash
GOOS=darwin GOARCH=amd64 go build -o sysara-darwin .
```

## 🐳 Docker Support

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o sysara .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/sysara .
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates
EXPOSE 8080
CMD ["./sysara"]
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

If you encounter any issues or have questions:

1. Check the [Issues](https://github.com/alpemreelmas/sysara/issues) page
2. Create a new issue with detailed information
3. Include system information and error logs

## 🗺️ Roadmap

- [ ] Multi-server support
- [ ] Docker container management
- [ ] Advanced alerting system
- [ ] REST API for external integrations
- [ ] Two-factor authentication
- [ ] Advanced user roles and permissions
- [ ] Database backups and restoration
- [ ] Plugin system for extensions

## 👥 Authors

- **alpemreelmas** - *Initial work* - [GitHub](https://github.com/alpemreelmas)

## 🙏 Acknowledgments

- [Gin Web Framework](https://gin-gonic.com/) for the excellent HTTP framework
- [GORM](https://gorm.io/) for the intuitive ORM
- [HTMX](https://htmx.org/) for seamless client-server interactions
- [Tailwind CSS](https://tailwindcss.com/) for the beautiful UI components
- [gopsutil](https://github.com/shirou/gopsutil) for system monitoring capabilities

---

**Sysara** - Empowering system administrators with modern, intuitive tools for server management.