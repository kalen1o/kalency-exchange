FROM node:22-alpine
WORKDIR /app

COPY package.json /app/package.json
COPY src /app/src

EXPOSE 8085
CMD ["npm", "start"]
