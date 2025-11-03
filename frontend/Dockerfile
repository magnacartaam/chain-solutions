# /frontend/Dockerfile

# --- Build Stage ---
# Use an official Node.js image to build our static assets.
FROM node:18-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy the package.json and package-lock.json files
COPY package*.json ./

# Install all dependencies
RUN npm install

# Copy the rest of the frontend source code
COPY . .

# Run the build command defined in package.json
# This will compile our Astro project into a 'dist' folder
RUN npm run build


# --- Final Stage ---
# Use a tiny, efficient Nginx server to serve the static files.
FROM nginx:1.25-alpine

# Copy the pre-built static files from the 'dist' folder of our builder stage
# into the default web root directory of the Nginx server.
COPY --from=builder /app/dist /usr/share/nginx/html

# Nginx will automatically serve the index.html file from this directory.
# Expose port 80 for the Nginx server to listen on.
EXPOSE 80