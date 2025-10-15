#!/bin/bash

# Vercel deployment script for Golang backend

echo "🚀 Deploying Golang Backend to Vercel..."

# Check if Vercel CLI is installed
if ! command -v vercel &> /dev/null; then
    echo "❌ Vercel CLI not found. Installing..."
    npm install -g vercel
fi

# Login to Vercel (if not already logged in)
echo "🔐 Checking Vercel authentication..."
vercel login --check || vercel login

# Deploy to Vercel
echo "📦 Deploying to Vercel..."
vercel --prod

echo "✅ Deployment complete!"
echo "🌐 Your API will be available at the URL shown above"
echo ""
echo "📝 Don't forget to:"
echo "   1. Set environment variables in Vercel dashboard:"
echo "      - MONGO_URI: Your MongoDB connection string"
echo "      - JWT_SECRET: A secure random string for JWT signing"
echo "      - ENCRYPTION_KEY: 32-byte key for AES encryption"
echo "   2. Update CORS settings if needed"
echo "   3. Configure MongoDB Atlas network access"
