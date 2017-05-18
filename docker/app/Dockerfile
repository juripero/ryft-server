FROM ubuntu:16.10

RUN apt-get update -y && apt-get install -y \
     ssh-client \
     dpkg \
     tar \
     less \
     vim \
     jq \
     curl && \
   apt-get autoremove -y && \
   apt-get clean

EXPOSE 8765

COPY .tmp/ryft-server /ryft-server
COPY etc/ryft-server.ryftprim.conf /etc/ryft-server.conf
COPY etc/ryft-users.yaml /etc/ryft-users.yaml
COPY ryftprim.ssh /usr/bin/ryftprim
COPY run.sh /run.sh

RUN groupadd ryftuser && \
    useradd -g ryftuser ryftuser && \
    chmod +x /ryft-server && \
    chmod +x /usr/bin/ryftprim && \
    chmod +x /run.sh && \
    mkdir -p /ryftone

CMD ['/bin/bash', '/run.sh']
