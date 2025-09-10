# Grade Processing System - MySQL Edition

A production-ready Golang backend system for processing and syncing student grades with an external Education Department API, using MySQL as the database.

## 🚀 Quick Start

### Prerequisites
- Docker and docker-compose
- Go 1.22+ (for local development)

### 1. Setup Project
```bash
# Create project directory and files
mkdir integration-education-db && cd integration-education-db

# Copy all the file contents from the artifacts above
# OR run the bash commands to create all files automatically

# Initialize Go modules
go mod init integration-education-db
go mod tidy
```

### 2. Start All Services
```bash
# Start all services (MySQL, Redis, MinIO, API, Workers)
docker-compose up --build
```

### 3. Verify System
```bash
# Check API health
curl http://localhost:8080/health

# Expected response:
{
  "status": "healthy",
  "service": "integration-education-db",
  "version": "1.0.0"
}
```

## 🗄️ Database Information

### **MySQL Configuration**
- **Host**: localhost:3306
- **Database**: grades_db
- **User**: root
- **Password**: rootpassword
- **Charset**: utf8mb4
- **Engine**: InnoDB

### **Database Schema**
```sql
-- Files table
CREATE TABLE files (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    s3_path VARCHAR(500) NOT NULL,
    school_id BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'UPLOADED',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Grades staging table
CREATE TABLE grades_staging (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    file_id BIGINT NOT NULL,
    student_id VARCHAR(50) NOT NULL,
    subject VARCHAR(100) NOT NULL,
    class VARCHAR(50) NOT NULL,
    semester VARCHAR(20) NOT NULL,
    grade DECIMAL(5,2) NOT NULL CHECK (grade >= 0 AND grade <= 100),
    status VARCHAR(20) NOT NULL DEFAULT 'READY',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);
```

## 🔧 MySQL-Specific Features

### **Optimizations**
- **Batch Inserts**: Efficient bulk operations for grade data
- **Proper Indexing**: Optimized queries for file and grade lookups
- **Foreign Key Constraints**: Data integrity enforcement
- **Auto-increment IDs**: MySQL-native primary keys
- **UTF8MB4 Charset**: Full Unicode support including emojis

### **Connection Configuration**
```yaml
database:
  host: mysql
  port: 3306
  user: root
  password: rootpassword
  name: grades_db
  charset: utf8mb4
  parse_time: true
  loc: UTC
  max_connections: 100
  max_idle_connections: 10
  connection_lifetime: 1h
```

## 🛠️ Development Commands

### **MySQL Management**
```bash
# Connect to MySQL CLI
make mysql-client

# View MySQL logs
make mysql-logs

# Run database migrations
make migrate-up

# Rollback migrations
make migrate-down
```

### **Development Workflow**
```bash
# Build all binaries
make build

# Run individual services locally
make dev-api          # Start API server
make dev-ingestion    # Start ingestion worker
make dev-sync         # Start sync worker

# Run tests
make test

# Format and lint code
make fmt
make lint
```

## 🌐 Service Endpoints

| Service | Port | Description |
|---------|------|-------------|
| **API Server** | 8080 | REST API endpoints |
| **MySQL** | 3306 | Database server |
| **Redis** | 6379 | Job queue and caching |
| **MinIO** | 9000 | S3-compatible storage |
| **MinIO Console** | 9001 | Web UI for MinIO |

### **API Endpoints**
- `GET /health` - Health check
- `POST /api/v1/sync/trigger` - Trigger grade sync
- `GET /api/v1/grades/status/{file_id}` - Get sync status

## 📊 Sample Usage

### **1. Trigger Sync**
```bash
curl -X POST http://localhost:8080/api/v1/sync/trigger \
  -H "Content-Type: application/json" \
  -d '{
    "file_id": 1,
    "class": "Grade-10",
    "subject": "Mathematics"
  }'
```

### **2. Check Status**
```bash
curl http://localhost:8080/api/v1/grades/status/1
```

### **3. Database Queries**
```sql
-- Check file status
SELECT * FROM files WHERE id = 1;

-- Check grade processing status
SELECT 
    status, 
    COUNT(*) as count 
FROM grades_staging 
WHERE file_id = 1 
GROUP BY status;

-- View failed grades with errors
SELECT 
    student_id, 
    error_message 
FROM grades_staging 
WHERE status = 'FAILED' AND file_id = 1;
```

## 🏗️ Architecture Highlights

### **MySQL-Specific Design**
1. **Efficient Batch Operations**: Uses multi-row INSERT statements
2. **Proper Indexing Strategy**: Optimized for common query patterns
3. **Foreign Key Relationships**: Ensures data integrity
4. **Charset Configuration**: UTF8MB4 for international support
5. **Transaction Management**: Ensures data consistency

### **Key Differences from PostgreSQL**
- Uses `?` placeholders instead of `$1, $2, ...`
- `AUTO_INCREMENT` instead of `SERIAL`
- `CURRENT_TIMESTAMP` instead of `NOW()`
- Different DSN format for connections
- MySQL-specific SQL syntax and functions

## 🚀 Scaling Considerations

### **MySQL Optimization**
```sql
-- Add indexes for heavy queries
CREATE INDEX idx_grades_staging_compound ON grades_staging(file_id, class, subject, status);

-- Optimize for bulk operations
SET SESSION bulk_insert_buffer_size = 256*1024*1024;
```

### **Horizontal Scaling**
```bash
# Scale workers independently
docker-compose up --scale ingestion-worker=5 --scale sync-worker=3
```

## 🔍 Troubleshooting

### **MySQL Connection Issues**
```bash
# Check MySQL service status
docker-compose ps mysql

# View MySQL logs
docker-compose logs mysql

# Test MySQL connection
docker-compose exec mysql mysql -u root -prootpassword -e "SELECT 1"
```

### **Common MySQL Errors**
1. **Connection refused**: Check if MySQL container is healthy
2. **Access denied**: Verify credentials in config.yaml
3. **Database doesn't exist**: Check if grades_db was created properly
4. **Charset issues**: Ensure UTF8MB4 is configured correctly

### **Performance Monitoring**
```sql
-- Check connection status
SHOW STATUS LIKE 'Connections';
SHOW STATUS LIKE 'Threads_connected';

-- Monitor slow queries
SHOW STATUS LIKE 'Slow_queries';

-- Check table sizes
SELECT 
    table_name,
    ROUND(((data_length + index_length) / 1024 / 1024), 2) AS 'Size (MB)'
FROM information_schema.tables 
WHERE table_schema = 'grades_db';
```

## 🔐 Security Features

### **MySQL Security**
- Root password protection
- Database-level access control  
- Connection encryption support
- SQL injection prevention with prepared statements

### **Application Security**
- Input validation and sanitization
- Structured error handling
- Secure credential management
- Request/response logging

## 📈 Monitoring & Observability

### **Health Checks**
- MySQL connectivity validation
- Redis queue monitoring  
- External API availability
- Worker process status

### **Logging Structure**
```json
{
  "level": "info",
  "time": "2024-01-15T10:30:00Z",
  "message": "Processing ingestion job",
  "file_id": 123,
  "s3_path": "uploads/grades.xlsx", 
  "mysql_connection_pool": "active:5,idle:3,max:100"
}
```

## 🎯 Production Checklist

### **MySQL Production Setup**
- [ ] Configure proper MySQL user (not root)
- [ ] Set up database backups
- [ ] Configure MySQL replication if needed
- [ ] Optimize MySQL configuration (my.cnf)
- [ ] Set up monitoring (slow query log, etc.)
- [ ] Configure proper resource limits

### **Application Production Setup**  
- [ ] Use environment variables for secrets
- [ ] Configure proper logging levels
- [ ] Set up health check endpoints
- [ ] Configure graceful shutdown
- [ ] Set resource limits in docker-compose
- [ ] Add metrics collection

---

## 🆚 Key Changes from PostgreSQL Version

| Component | PostgreSQL | MySQL |
|-----------|------------|-------|
| **Driver** | `lib/pq` | `go-sql-driver/mysql` |
| **DSN Format** | `host=... port=... user=...` | `user:pass@tcp(host:port)/db?params` |
| **Placeholders** | `$1, $2, $3` | `?, ?, ?` |
| **Auto Increment** | `SERIAL` | `AUTO_INCREMENT` |
| **Timestamps** | `TIMESTAMP WITH TIME ZONE` | `TIMESTAMP` |
| **NOW()** | `NOW()` | `CURRENT_TIMESTAMP` |
| **Array Operations** | `= ANY($1)` | `IN (?, ?, ?)` |
| **Default Port** | 5432 | 3306 |

The system maintains the same functionality and architecture while being optimized for MySQL's specific features and syntax.
