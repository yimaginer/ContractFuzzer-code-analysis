# 使用官方的 Go 语言 Alpine 镜像作为基础镜像
FROM golang:alpine

# 安装必要的工具和依赖
RUN \
  apk add --update git make gcc musl-dev linux-headers  # 安装 Git、Make、GCC 和其他开发工具

# 设置 Node.js 的版本
ENV NODE_VERSION 9.11.1

# 添加 Node.js 环境
RUN addgroup -g 1000 node \ # 创建 node 用户组
    && adduser -u 1000 -G node -s /bin/sh -D node \ # 创建 node 用户
    && apk add --no-cache \
        libstdc++ \ # 安装 C++ 标准库
    && apk add --no-cache --virtual .build-deps \ # 安装构建 Node.js 所需的依赖
        binutils-gold \
        curl \
        g++ \
        gcc \
        gnupg \
        libgcc \
        linux-headers \
        make \
        python \
  # 导入 Node.js 的 GPG 密钥
  && for key in \
    94AE36675C464D64BAFA68DD7434390BDBE9B9C5 \
    FD3A5288F042B6850C66B31F09FE44734EB7990E \
    71DCFD284A79C3B38668286BC97EC7A07EDE3FC1 \
    DD8F2338BAE7501E3DD5AC78C273792F7D83545D \
    C4F0DFFF4E8C1A8236409D08E73BC641CC11F4C8 \
    B9AE9905FFD7803F25714661B63B535A4C206CA9 \
    56730D5401028683275BD23C23EFEFE93C4CFFFE \
    77984A986EBC2AA786BC0F66B01FBB92821C587A \
  ; do \
    gpg --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys "$key" || \
    gpg --keyserver hkp://ipv4.pool.sks-keyservers.net --recv-keys "$key" || \
    gpg --keyserver hkp://pgp.mit.edu:80 --recv-keys "$key" ; \
  done \
    # 下载并安装 Node.js
    && curl -SLO "https://nodejs.org/dist/v$NODE_VERSION/node-v$NODE_VERSION.tar.xz" \
    && curl -SLO --compressed "https://nodejs.org/dist/v$NODE_VERSION/SHASUMS256.txt.asc" \
    && gpg --batch --decrypt --output SHASUMS256.txt SHASUMS256.txt.asc \
    && grep " node-v$NODE_VERSION.tar.xz\$" SHASUMS256.txt | sha256sum -c - \
    && tar -xf "node-v$NODE_VERSION.tar.xz" \
    && cd "node-v$NODE_VERSION" \
    && ./configure \
    && make -j$(getconf _NPROCESSORS_ONLN) \
    && make install \
    && apk del .build-deps \ # 删除构建依赖以减小镜像大小
    && cd .. \
    && rm -Rf "node-v$NODE_VERSION" \
    && rm "node-v$NODE_VERSION.tar.xz" SHASUMS256.txt.asc SHASUMS256.txt

# 设置 Yarn 的版本
ENV YARN_VERSION 1.5.1

# 安装 Yarn
RUN apk add --no-cache --virtual .build-deps-yarn curl gnupg tar \
  && for key in \
    6A010C5166006599AA17F08146C2130DFD2497F5 \
  ; do \
    gpg --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys "$key" || \
    gpg --keyserver hkp://ipv4.pool.sks-keyservers.net --recv-keys "$key" || \
    gpg --keyserver hkp://pgp.mit.edu:80 --recv-keys "$key" ; \
  done \
  && curl -fSLO --compressed "https://yarnpkg.com/downloads/$YARN_VERSION/yarn-v$YARN_VERSION.tar.gz" \
  && curl -fSLO --compressed "https://yarnpkg.com/downloads/$YARN_VERSION/yarn-v$YARN_VERSION.tar.gz.asc" \
  && gpg --batch --verify yarn-v$YARN_VERSION.tar.gz.asc yarn-v$YARN_VERSION.tar.gz \
  && mkdir -p /opt \
  && tar -xzf yarn-v$YARN_VERSION.tar.gz -C /opt/ \
  && ln -s /opt/yarn-v$YARN_VERSION/bin/yarn /usr/local/bin/yarn \
  && ln -s /opt/yarn-v$YARN_VERSION/bin/yarnpkg /usr/local/bin/yarnpkg \
  && rm yarn-v$YARN_VERSION.tar.gz.asc yarn-v$YARN_VERSION.tar.gz \
  && apk del .build-deps-yarn \
  && apk add --no-cache git

# 安装 Babel CLI
RUN yarn global add babel-cli \
    && ln -s /usr/bin/babel-node /usr/bin/bnode

# 创建 ContractFuzzer 工作目录
RUN mkdir -p /ContractFuzzer 

# 设置工作目录
WORKDIR /ContractFuzzer

# 添加项目文件到镜像中
ADD go-ethereum-cf go-ethereum
ADD Ethereum Ethereum
ADD examples examples
ADD contract_fuzzer contract_fuzzer
ADD contract_tester contract_tester

# 添加运行脚本到镜像中
ADD fuzzer_run.sh fuzzer_run.sh
ADD tester_run.sh tester_run.sh
ADD geth_run.sh  geth_run.sh
ADD run.sh  run.sh

# 构建项目
RUN \
  (cd go-ethereum && make geth)                                && \ # 构建 Geth
  (cd contract_fuzzer && source ./gopath.sh && cd ./src/ContractFuzzer/contractfuzzer && go build -o contract_fuzzer) && \ # 构建 ContractFuzzer
  cp contract_fuzzer/src/ContractFuzzer/contractfuzzer/contract_fuzzer /usr/local/bin   && \ # 将 ContractFuzzer 可执行文件复制到系统路径
  cp go-ethereum/build/bin/geth /usr/local/bin/                && \ # 将 Geth 可执行文件复制到系统路径
  apk del git  make gcc musl-dev linux-headers                 && \ # 删除不再需要的依赖
  rm -rf ./go-ethereum && rm -rf ./contract_fuzzer                 && \ # 删除构建目录
  rm -rf /var/cache/apk/*                    # 清理缓存

# 设置默认命令
CMD ["sh"]
