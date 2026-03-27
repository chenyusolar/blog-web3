# AIGen-Blog (Web3 + AI) 🚀

A modern, powerful blog platform powered by **Go (Hertz)**, **Vue 3**, and **Solana Web3**.

This platform combines AI-assisted content creation with a tokenized reward economy, empowering users to earn `BLOG` tokens through their activity.

## ✨ Key Features

### 🤖 AI Content Creation
- **AI-Driven Writing**: Integrated with AI services (Eino, GPT-4o-mini) to generate blog content from prompts with consistent quality.
- **Smart Formatting**: Automatic categorizing and tagging of generated posts.

### 🔗 Web3 & Rewards (Solana)
- **Automatic Wallet Deployment**: Every new user automatically receives a Solana wallet address and private key (delivered securely via email).
- **Activity Rewards**: Earn `BLOG` tokens for:
    - **Publishing**: Gain tokens for each new post.
    - **Engaging**: Rewards for comments and interactions.
    - **Forwarding**: Spread content across the platform to earn "Forward" rewards.
- **3-Level Referral System**: A recursive reward mechanism that benefits inviters up to three tiers deep.

### 🛠️ Administrative Dashboard
- **Platform Analytics**: Real-time stats on users, blog counts, and total token issuance.
- **User Management**: Comprehensive oversight of user balances and wallet activities.
    - View and filter all registered users.
    - Monitor `BLOG` token distribution history (`RewardLog`).
- **Live System Config**: Update platform parameters (Reward amounts, referral rates, Solana RPC) directly from the UI.

### 📱 Premium UX/UI
- **Responsive Design**: Built with Vue 3 + Tailwind CSS, optimized for all devices.
- **My Wallet View**: A dashboard to track earnings, referral details, and transaction history.
- **Interactive UI**: Elegant use of `lucide-vue-next` icons and smooth transitions.

---

## 🛠️ Tech Stack

### Backend
- **Core**: Go (Hertz Framework)
- **Database**: GORM (PostgreSQL) + Redis (Caching)
- **Web3**: Solana Go SDK (for wallet generation and token logic)
- **Security**: JWT-based authentication, bcrypt password hashing.

### Frontend
- **Framework**: Vue 3 (Composition API)
- **State Management**: Pinia (Store-based architecture)
- **Styling**: Tailwind CSS (Premium glassmorphism and modern palettes)
- **Routing**: Vue Router (Protected routes for Admin and Wallet)
- **Communication**: Axios (with centralized interceptors for Token management)

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
REWARD_FORWARD=2.0
REF_L1=10.0
REF_L2=5.0
REF_L3=2.0

# Email Delivery (SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=...
SMTP_PASS=...
EMAIL_FROM=AIGen-Blog <noreply@aigenblog.com>
```

### 2. Backend Setup
```bash
cd backend
go mod tidy
go build -o backend.exe .
./backend.exe
```

### 3. Frontend Setup
```bash
cd frontend
npm install
npm run dev
```

---

## 🛡️ Security
- **Custodial-Lite**: Users receive their private keys upon registration. **Important:** Users are advised to move significant funds to non-custodial wallets.
- **Admin Access**: `/admin` routes are strictly guarded by role-based AuthMiddleware.

## 📄 License
MIT License. Created by the AIGen-Blog Team.
