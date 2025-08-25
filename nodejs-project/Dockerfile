FROM node:18-alpine

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm i

# Copy source code
COPY . .

# Expose port
EXPOSE 5555

# Start the application
CMD ["npm", "start"]
