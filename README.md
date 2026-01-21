# ☁️ Cloud Guard SaaS (MVP)

> **Auto-detect wasted resources, security risks, and cost anomalies in your AWS accounts.**

Cloud Guard is a lightweight, read-only SaaS tool designed for startups and indie hackers to keep their AWS bills low and security high. It connects securely via IAM Roles (no long-term credentials) and provides a simple dashboard and Slack alerts.

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/docker-%230blue.svg?style=for-the-badge&logo=docker&logoColor=white)
![AWS](https://img.shields.io/badge/AWS-%23FF9900.svg?style=for-the-badge&logo=amazon-aws&logoColor=white)
![SQLite](https://img.shields.io/badge/sqlite-%2307405e.svg?style=for-the-badge&logo=sqlite&logoColor=white)

## ✨ Features

- **🛡️ Secure Connection**: Uses `sts:AssumeRole` with ReadOnly permissions. No Access Keys stored.
- **💸 Cost Scanner**: Weekly cost summaries and breakdown by service via Cost Explorer.
- **🖥️ EC2 Scanner**: 
  - Detects **Stopped Instances** that are just sitting there.
  - Detects **Idle Instances** (Avg CPU < 5% over 7 days).
- **🗄️ S3 Scanner**:
  - Detects **Public Buckets** (via Policy checks).
  - Checks for missing Public Access Blocks.
- **🐘 RDS Scanner**:
  - Detects **Over-provisioned Databases** (Large instance types with low CPU).
- **🚨 Alerts**: Real-time notifications to Slack.
- **📊 Minimal Dashboard**: View all findings and cost data in one place.

## 🚀 Getting Started

### Prerequisites

- [Go 1.23+](https://golang.org/dl/)
- AWS Account (to scan)
- Slack Webhook URL (optional, for alerts)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/singh-anurag-7991/cloud-guard.git
   cd cloud-guard
   ```

2. **Run Locally**
   ```bash
   # Set Slack Webhook (Optional)
   export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK"
   
   # Run the server
   go run cmd/server/main.go
   ```
   The dashboard will be available at [http://localhost:8080](http://localhost:8080).

### 🐳 Docker Usage

You can also run Cloud Guard using Docker:

```bash
docker build -t cloud-guard .
docker run -p 8080:8080 -e SLACK_WEBHOOK_URL="your_webhook_url" cloud-guard
```

## ⚙️ Configuration: Connecting AWS

To scan an AWS account, you must grant Cloud Guard permission via an IAM Role.

1. **Create an IAM Role** in the target AWS account.
2. **Select "Custom Trust Policy"** and use the following JSON:
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Principal": {
           "AWS": "arn:aws:iam::YOUR_HOSTING_ACCOUNT_ID:root"
         },
         "Action": "sts:AssumeRole"
       }
     ]
   }
   ```
   *Replace `YOUR_HOSTING_ACCOUNT_ID` with the Account ID where Cloud Guard is running. For local testing, use your personal User ARN.*

3. **Attach Permissions**: Add the `ReadOnlyAccess` (or `ViewOnlyAccess`) managed policy.
4. **Copy the Role ARN** (e.g., `arn:aws:iam::123456789012:role/CloudGuardReader`).
5. **Paste the ARN** into the Cloud Guard Dashboard to start scanning.

## 🧪 Testing

Running the unit tests and benchmarks:

```bash
go test ./... -v      # Run all tests
go test ./... -bench=. # Run benchmarks
```

### Benchmark Results (Example)
The Rule Engine is optimized for speed.
```text
BenchmarkRuleEngine-10    	 2528148	       468.3 ns/op
```
*(Performance may vary based on hardware)*

## 🏗️ Architecture

- **Backend**: Go (Standard Library + AWS SDK v2).
- **Database**: SQLite (Embedded, zero-config).
- **Frontend**: Server-side rendered HTML (Standard `html/template`).
- **Orchestrator**: Manages scanning jobs and rule evaluation.
- **Scanners**: Independent modules for each AWS service.

## 📝 Roadmap

- [ ] Email reports (SendGrid/SES).
- [ ] Multi-user support with Auth0/Cognito.
- [ ] Configurable thresholds (e.g., change "Idle" from 5% to 10%).
- [ ] Auto-remediation (e.g., stop idle EC2 instances).

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.
