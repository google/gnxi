FROM node:14-alpine AS dev

WORKDIR /workspace

COPY . .

RUN npm cache verify && npm i && npm run build

CMD [ "npm", "start" ]


FROM nginx:alpine

WORKDIR /etc/nginx/conf.d

COPY default.conf .

WORKDIR /usr/share/nginx/html

COPY --from=dev /workspace/dist/webui .
