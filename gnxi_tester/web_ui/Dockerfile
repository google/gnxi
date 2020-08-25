FROM node:14-alpine AS dev

WORKDIR /workspace

COPY . .

ARG BACKEND_HOST

RUN echo "export const environment = { apiUrl: '$BACKEND_HOST' };" > ./src/app/environment.ts

RUN npm i && npm run-script build

CMD [ "npm", "start" ]


FROM nginx:alpine

WORKDIR /usr/share/nginx/html

COPY --from=dev /workspace/dist/webui .
