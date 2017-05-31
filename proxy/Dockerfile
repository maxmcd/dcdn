FROM node:boron

RUN mkdir -p /opt/app/
WORKDIR /opt/app/
VOLUME /opt/app/

COPY package.json /opt/app/
RUN npm install

COPY . /opt/app/