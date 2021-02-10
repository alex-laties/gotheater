FROM golang:1.15 as gobuild
WORKDIR /usr/src/app
COPY . .
RUN cd cmd/gotheater && go build

# FROM node:15 as jsbuild
# WORKDIR /usr/src/app
# COPY frontend/package*.json .
# RUN npm install
# COPY frontend/ .
# RUN npm run build

FROM debian:buster-slim
RUN mkdir -p /var/lib/gotheater /etc/gotheater
COPY --from=gobuild /usr/src/app/cmd/gotheater/gotheater /usr/local/bin/gotheater
# COPY --from=jsbuild /usr/src/app/build /var/lib/gotheater/frontend
# COPY ./conf /etc/gotheater

CMD ["gotheater"]