# AIGen-Blog (Web3 + AI) 🚀

A modern, powerful blog platform powered by **Go (Hertz)**, **Vue 3**, and **Solana Web3**.

This platform combines AI-assisted content creation with a tokenized reward economy, empowering users to earn `BLOG` tokens through their activity and influence.

## ✨ Key Features

### 🤖 AI Content Creation
- **AI-Driven Writing**: Integrated with AI services (Eino, GPT-4o-mini) to generate blog content from prompts with consistent quality.
- **Premium Tiptap Editor**: A rich editing experience with support for:
    - YouTube and Image embedding.
    - Advanced Table management.
    - Secure File Attachments & Uploads.
    - Modern Toolbar with glassmorphism design.

### 🔗 Web3 & Rewards (Solana)
- **Automatic Wallet Deployment**: Every new user automatically receives a Solana wallet address and private key (delivered securely via email).
- **Social Share-to-Earn**: Earn `BLOG` tokens by sharing articles to external platforms (**Twitter/X, Telegram, Facebook**). 
    - Shared links automatically include your **Referral Code**.
    - Track global `ShareCount` for every article.
- **Activity Rewards**: Earn tokens for Publishing, Engaging (comments), and Sharing.
- **3-Level Referral System**: A recursive reward mechanism that benefits inviters up to three tiers deep.

### 🔍 Discovery & UX
- **Real-time Search**: Instant filtering of articles by title or author directly from the homepage.
- **Responsive Design**: Built with Vue 3 + Tailwind CSS, optimized for all devices.
- **My Wallet View**: A dashboard to track earnings, referral details, and transaction history.

### 🛠️ Administrative Dashboard
- **Platform Analytics**: Real-time stats on users, blog counts, and total token issuance.
- **User Management**: Comprehensive oversight of user balances and wallet activities.
- **Live System Config**: Update platform parameters (Reward amounts, referral rates, Solana RPC) directly from the UI.

---

## 🛠️ Tech Stack

### Backend
- **Core**: Go (Hertz Framework)
- **Database**: GORM (PostgreSQL) + Redis (Caching)
- **Web3**: Solana Go SDK (for wallet generation and token logic)
- **Security**: JWT-based authentication, bcrypt password hashing.

### Frontend
- **Framework**: Vue 3 (Composition API)
- **State Management**: Pinia
- **Styling**: Tailwind CSS (Premium glassmorphism and modern palettes)
- **Components**: Lucide Icons, Tiptap Editor.

---

## 🚀 Getting Started

### Prerequisites
- [Go 1.22+](https://go.dev/)
- [Node.js 18+](https://nodejs.org/)
- [PostgreSQL](https://www.postgresql.org/) & [Redis](https://redis.io/)

### 1. Environment Configuration
Create a `.env` file in the `backend/` directory:

```env
PORT=8888
DB_DSN=host=localhost user=postgres password=123456 dbname=postgres port=5432 sslmode=disable
REDIS_ADDR=localhost:6379

# AI Configuration
AI_API_KEY=your_actual_api_key_here
AI_BASE_URL=https://api.openai.com/v1
AI_MODEL=gpt-4o-mini

# Web3 Configuration
SOLANA_RPC=https://api.devnet.solana.com
ADMIN_WALLET_ADDRESS=...
ADMIN_WALLET_KEY=...

# Reward Amounts
REWARD_POST=5.0
REWARD_COMMENT=1.0
REWARD_FORWARD=2.0  # Social Share Reward
REF_L1=10.0
REF_L2=5.0
REF_L3=2.0

# Email Delivery (SMTP)
...
```

### 2. Setup & Run
```bash
# Backend
cd backend && go mod tidy && go run .

# Frontend
cd frontend && npm install && npm run dev
```

---

## 📄 License
MIT License. Created by the AIGen-Blog Team.
