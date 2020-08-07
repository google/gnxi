FROM node:14-alpine AS dev

WORKDIR /workspace

COPY . .

RUN npm i && npm build

CMD [ "npm", "start" ]


FROM nginx:alpine

WORKDIR /usr/share/nginx/html

COPY --from=dev /workspace/dist/webui .
