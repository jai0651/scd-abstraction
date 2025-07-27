# Setup Guide - SCD Abstraction Project

This guide provides detailed instructions for setting up the SCD (Slowly Changing Dimension) abstraction project on your local machine.

## 📋 Prerequisites

### Required Software
- **Docker & Docker Compose** - For containerized database and services
- **Python 3.11+** - For template generation and Python benchmarks
- **Go 1.24+** - For Go benchmarks and development
- **Git** - For version control

### System Requirements
- **RAM**: Minimum 4GB (8GB recommended)
- **Storage**: At least 2GB free space
- **OS**: macOS, Linux, or Windows (with WSL2 for Windows)

## 🚀 Quick Setup (Recommended)

### 1. Clone the Repository
```bash
git clone <repository-url>
cd scd-abstraction
```

### 2. Start Database
```bash
# Start PostgreSQL database
docker-compose up -d db

# Verify database is running
docker-compose ps
```

### 3. Generate SCD Helpers
```bash
# Install Python dependencies (if needed)
pip install jinja2

# Generate for Go/GORM
python generate_scd_helper.py gorm

# Generate for Python/Django
python generate_scd_helper.py django
```

### 4. Run Go Benchmarks
```bash
# Run Go benchmarks in Docker
docker-compose up --build go-bench
```

## 🔧 Detailed Setup Instructions

### Step 1: Environment Preparation

#### Install Docker & Docker Compose

**macOS:**
```bash
# Using Homebrew
brew install --cask docker

# Or download from Docker Desktop
# https://www.docker.com/products/docker-desktop
```

**Linux (Ubuntu/Debian):**
```bash
# Update package index
sudo apt-get update

# Install prerequisites
sudo apt-get install apt-transport-https ca-certificates curl gnupg lsb-release

# Add Docker's official GPG key
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# Set up stable repository
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine
sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Add user to docker group
sudo usermod -aG docker $USER

# Start Docker service
sudo systemctl start docker
sudo systemctl enable docker
```

**Windows:**
1. Download Docker Desktop from https://www.docker.com/products/docker-desktop
2. Install and restart your computer
3. Enable WSL2 if prompted

#### Install Python 3.11+

**macOS:**
```bash
# Using Homebrew
brew install python@3.11

# Verify installation
python3 --version
```

**Linux (Ubuntu/Debian):**
```bash
# Add deadsnakes PPA
sudo add-apt-repository ppa:deadsnakes/ppa
sudo apt update

# Install Python 3.11
sudo apt install python3.11 python3.11-pip python3.11-venv

# Create symlink
sudo ln -s /usr/bin/python3.11 /usr/bin/python3

# Verify installation
python3 --version
```

**Windows:**
1. Download Python 3.11+ from https://www.python.org/downloads/
2. Install with "Add Python to PATH" checked
3. Verify installation: `python --version`

#### Install Go 1.24+

**macOS:**
```bash
# Using Homebrew
brew install go

# Verify installation
go version
```

**Linux:**
```bash
# Download Go
wget https://go.dev/dl/go1.24.linux-amd64.tar.gz

# Extract to /usr/local
sudo tar -C /usr/local -xzf go1.24.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

**Windows:**
1. Download Go from https://go.dev/dl/
2. Install and restart your computer
3. Verify installation: `go version`

### Step 2: Project Setup

#### Clone and Navigate
```bash
# Clone the repository
git clone <repository-url>
cd scd-abstraction

# Verify project structure
ls -la
```

Expected output:
```
docker-compose.yml
generate_scd_helper.py
scd_helper_template.jinja
Go/
python/
README.md
setup.md
```

#### Install Python Dependencies
```bash
# Create virtual environment (recommended)
python3 -m venv venv

# Activate virtual environment
# On macOS/Linux:
source venv/bin/activate
# On Windows:
venv\Scripts\activate

# Install required packages
pip install jinja2

# Verify installation
python -c "import jinja2; print('Jinja2 installed successfully')"
```

### Step 3: Database Setup

#### Start PostgreSQL Database
```bash
# Start the database service
docker-compose up -d db

# Wait for database to be ready
sleep 10

# Verify database is running
docker-compose ps
```

Expected output:
```
NAME                COMMAND                  SERVICE             STATUS              PORTS
scd-abstraction-db-1   "docker-entrypoint.s…"   db                  running             0.0.0.0:5432->5432/tcp
```

#### Verify Database Connection
```bash
# Test database connection
docker-compose exec db psql -U postgres -d scd -c "SELECT version();"
```

Expected output:
```
PostgreSQL 15.x on x86_64-pc-linux-gnu, compiled by gcc (GCC) 11.2.0, 64-bit
```

### Step 4: Generate SCD Helpers

#### Generate Go/GORM Helpers
```bash
# Generate SCD helpers for Go/GORM
python generate_scd_helper.py gorm

# Verify generation
ls -la Go/scd/
```

Expected output:
```
scd_helpers.go
```

#### Generate Python/Django Helpers
```bash
# Generate SCD helpers for Python/Django
python generate_scd_helper.py django

# Verify generation
ls -la python/django_example/scd/
```

Expected output:
```
scd_helpers.py
```

### Step 5: Go Setup and Testing

#### Install Go Dependencies
```bash
# Navigate to Go directory
cd Go

# Install dependencies
go mod tidy

# Verify dependencies
go list -m all
```

#### Run Go Migrations
```bash
# Run migrations and seed data
go run main.go
```

Expected output:
```
Database connected successfully
Migrations completed
Data seeded successfully
```

#### Test Go Implementation
```bash
# Run Go tests
go test ./...

# Run specific SCD tests
go test ./repos/...
```

### Step 6: Python Setup and Testing

#### Install Python Dependencies
```bash
# Navigate to Python directory
cd ../python

# Install Django and PostgreSQL adapter
pip install django psycopg2-binary

# Verify installation
python -c "import django; print(f'Django {django.get_version()} installed')"
```

#### Run Django Migrations
```bash
# Run Django migrations
python manage.py migrate

# Verify migrations
python manage.py showmigrations
```

#### Test Python Implementation
```bash
# Run Django tests
python manage.py test

# Run SCD-specific tests
python django_example/test_scd.py
```

### Step 7: Run Benchmarks

#### Go Benchmarks
```bash
# Navigate back to project root
cd ..

# Run Go benchmarks in Docker
docker-compose up --build go-bench
```

Expected output:
```
=== Simple SCD Impact Analysis ===
Testing with 10,000 records

Job_By_Company/SCD_Abstraction-8         1000        1234567 ns/op
Job_By_Company/Raw_SQL_Equivalent-8      1000        1098765 ns/op

Job_By_Contractor/SCD_Abstraction-8      1000        1154321 ns/op
Job_By_Contractor/Raw_SQL_Equivalent-8   1000        1023456 ns/op
```

#### Python Benchmarks (Optional)
```bash
# Run Python benchmarks in Docker
docker-compose up --build python-bench
```

### Step 8: Verify Setup

#### Check Generated Files
```bash
# Verify Go helpers
head -20 Go/scd/scd_helpers.go

# Verify Python helpers
head -20 python/django_example/scd/scd_helpers.py
```

#### Test Database Connection
```bash
# Test Go database connection
cd Go
go run main.go

# Test Python database connection
cd ../python
python manage.py shell -c "from django.db import connection; print('Database connected successfully')"
```

## 🔧 Configuration Options

### Database Configuration

#### Using Different Database
Edit `docker-compose.yml`:
```yaml
services:
  db:
    image: postgres:15
    environment:
      POSTGRES_DB: scd
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
```

#### Using External Database
Edit `Go/main.go`:
```go
dsn := "host=localhost user=your_user password=your_password dbname=your_db port=5432 sslmode=disable"
```

Edit `python/myproject/settings.py`:
```python
DATABASES = {
    'default': {
        'ENGINE': 'django.db.backends.postgresql',
        'NAME': 'your_db',
        'USER': 'your_user',
        'PASSWORD': 'your_password',
        'HOST': 'localhost',
        'PORT': '5432',
    }
}
```

### Customizing SCD Generation

#### Adding New ORM Support
1. Edit `scd_helper_template.jinja`:
```jinja
{% elif target == "your_orm" %}
// Your ORM implementation
function latestSubquery(model) {
    return model.select('id, MAX(version) as max_version')
                .groupBy('id');
}
{% endif %}
```

2. Edit `generate_scd_helper.py`:
```python
elif target == 'your_orm':
    output_file = os.path.join('your_language', 'scd', 'scd_helpers.py')
```

3. Generate helpers:
```bash
python generate_scd_helper.py your_orm
```

## 🐛 Troubleshooting

### Common Issues

#### Docker Issues
```bash
# Check Docker status
docker --version
docker-compose --version

# Restart Docker service
sudo systemctl restart docker

# Check container logs
docker-compose logs db
```

#### Database Connection Issues
```bash
# Check if database is running
docker-compose ps

# Check database logs
docker-compose logs db

# Restart database
docker-compose restart db
```

#### Go Issues
```bash
# Check Go version
go version

# Clean Go cache
go clean -cache

# Update Go modules
go mod tidy
```

#### Python Issues
```bash
# Check Python version
python3 --version

# Reinstall virtual environment
rm -rf venv
python3 -m venv venv
source venv/bin/activate
pip install jinja2 django psycopg2-binary
```

#### Permission Issues
```bash
# Fix Docker permissions (Linux)
sudo usermod -aG docker $USER
newgrp docker

# Fix file permissions
chmod +x generate_scd_helper.py
```

### Performance Issues

#### Database Performance
```bash
# Check database performance
docker-compose exec db psql -U postgres -d scd -c "SELECT * FROM pg_stat_activity;"

# Optimize database
docker-compose exec db psql -U postgres -d scd -c "VACUUM ANALYZE;"
```

#### Benchmark Performance
```bash
# Run with smaller dataset
# Edit Go/benchmark/focused_scd_impact_test.go
# Change record count from 10000 to 1000

# Run specific benchmark
go test -bench=BenchmarkSimpleSCDImpact -benchtime=1s ./benchmark
```

## 📚 Next Steps

### Explore the Codebase
```bash
# Review generated SCD helpers
cat Go/scd/scd_helpers.go
cat python/django_example/scd/scd_helpers.py

# Review benchmarks
cat Go/benchmark/focused_scd_impact_test.go
```

### Run Custom Tests
```bash
# Test specific functionality
cd Go
go test -v ./repos/...

# Test Python functionality
cd ../python
python manage.py test django_example
```

### Extend the Project
```bash
# Add new ORM support
# Edit scd_helper_template.jinja and generate_scd_helper.py

# Add new benchmarks
# Create new test files in Go/benchmark/

# Add new models
# Create new files in Go/models/ and python/django_example/
```

## ✅ Verification Checklist

- [ ] Docker and Docker Compose installed
- [ ] Python 3.11+ installed
- [ ] Go 1.24+ installed
- [ ] Repository cloned
- [ ] Database started (`docker-compose up -d db`)
- [ ] SCD helpers generated (`python generate_scd_helper.py gorm django`)
- [ ] Go dependencies installed (`go mod tidy`)
- [ ] Go migrations run (`go run main.go`)
- [ ] Python dependencies installed (`pip install django psycopg2-binary`)
- [ ] Django migrations run (`python manage.py migrate`)
- [ ] Go tests pass (`go test ./...`)
- [ ] Python tests pass (`python manage.py test`)
- [ ] Benchmarks run successfully (`docker-compose up --build go-bench`)

## 🆘 Getting Help

### Documentation
- [README.md](README.md) - Project overview and architecture
- [SCD_Impact_Analysis.md](docs/SCD_Impact_Analysis.md) - Performance analysis
- [Template_vs_Alternatives_Analysis.md](docs/Template_vs_Alternatives_Analysis.md) - Architectural comparison

### Common Commands
```bash
# Restart everything
docker-compose down
docker-compose up -d db
python generate_scd_helper.py gorm django
docker-compose up --build go-bench

# Clean and rebuild
docker-compose down -v
docker system prune -f
docker-compose up -d db
```

### Support
- Check the troubleshooting section above
- Review Docker logs: `docker-compose logs`
- Check database logs: `docker-compose logs db`
- Verify file permissions and paths

---

*Setup completed successfully! You can now explore the SCD abstraction implementation and run benchmarks.* 