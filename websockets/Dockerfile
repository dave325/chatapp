# Create image based on the official Node image from dockerhub
FROM node:lts-alpine
# Create a directory where our app will be placed
RUN mkdir /code
# Change directory so that our commands run inside this new directory
WORKDIR /code
# Copy dependency definitions
COPY package.json /code
# Install dependecies
RUN npm install

RUN npm install -g @angular/cli
EXPOSE 4200 49153
# Get all the code needed to run the app
COPY . /code
