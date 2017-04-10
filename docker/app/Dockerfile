FROM ubuntu:16.10

RUN apt-get update -y && apt-get install -y \
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
COPY etc/ryft-server.conf /etc/ryft-server.conf
COPY etc/ryft-users.yaml /etc/ryft-users.yml
COPY ryftprim /usr/bin/ryftprim

RUN groupadd ryftuser && \
    useradd -g ryftuser ryftuser && \
    chmod +x /ryft-server && \
    chmod +x /usr/bin/ryftprim && \
    mkdir -p /ryftone

CMD /ryft-server --config=/etc/ryft-server.conf --address :8765 --debug