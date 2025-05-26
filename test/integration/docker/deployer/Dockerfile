FROM golang:1.24

WORKDIR /

RUN apt-get update --fix-missing && apt-get install -y ca-certificates openssh-client curl gnupg docker.io

RUN install -m 0755 -d /etc/apt/keyrings
RUN curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
RUN chmod a+r /etc/apt/keyrings/docker.gpg
RUN echo \
    "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
    "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
    tee /etc/apt/sources.list.d/docker.list > /dev/null

RUN apt-get update --fix-missing && apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

COPY *.sh .
COPY app/ app/

RUN chmod +x ./boot.sh
RUN chmod +x ./setup.sh

RUN rm -rf /root/.ssh
RUN ln -s /shared/ssh /root/.ssh
RUN mkdir -p /etc/docker/certs.d/registry:4443 && ln -s /shared/certs/domain.crt /etc/docker/certs.d/registry:4443/ca.crt

RUN git config --global user.email "deployer@faino.dev"
RUN git config --global user.name "Deployer"
RUN cd app && git init && git add . && git commit -m "feat: initial commit"
RUN cd app && git rev-parse --short HEAD > version.txt

HEALTHCHECK --interval=1s CMD pgrep sleep

CMD ["./boot.sh"]
