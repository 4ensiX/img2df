Convert docker container image to Dockerfile<br>
v0.0.1<br><br>

- local docker coniner image to Dockerfile
- install docker

usage: img2df [image name] or [image:tag]<br>
example: img2df debian
```
$ sudo ./img2df python:alpine
$ cat Dockerfile
FROM scratch

ADD file:aad4290d27580cc1a094ffaf98c3ca2fc5d699fe695dfb8e6e9fac20f1129450  /

CMD ["/bin/sh"]

ENV PATH=/usr/local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

ENV LANG=C.UTF-8

RUN set -eux;\
apk add --no-cache\
        ca-certificates\
        tzdata\
;

ENV GPG_KEY=A035C8C19219BA821ECEA86B64E628F8D684696D

ENV PYTHON_VERSION=3.10.0

RUN set -ex\
&& apk add --no-cache --virtual .fetch-deps\
        gnupg\
        tar\
        xz\
&& wget -O python.tar.xz "https://www.python.org/ftp/python/${PYTHON_VERSION%%[a-z]*}/Python-$PYTHON_VERSION.tar.xz"\
&& wget -O python.tar.xz.asc "https://www.python.org/ftp/python/${PYTHON_VERSION%%[a-z]*}/Python-$PYTHON_VERSION.tar.xz.asc"\
&& export GNUPGHOME="$(mktemp -d)"\
&& gpg --batch --keyserver hkps://keys.openpgp.org --recv-keys "$GPG_KEY"\
&& gpg --batch --verify python.tar.xz.asc python.tar.xz\
&& { command -v gpgconf > /dev/null && gpgconf --kill all || :; }\
&& rm -rf "$GNUPGHOME" python.tar.xz.asc\
&& mkdir -p /usr/src/python\
&& tar -xJC /usr/src/python --strip-components=1 -f python.tar.xz\
&& rm python.tar.xz\
&& apk add --no-cache --virtual .build-deps \
        bluez-dev\
        bzip2-dev\
        coreutils\
        dpkg-dev dpkg\
        expat-dev\
        findutils\
        gcc\
        gdbm-dev\
        libc-dev\
        libffi-dev\
        libnsl-dev\
        libtirpc-dev\
        linux-headers\
        make\
        ncurses-dev\
        openssl-dev\
        pax-utils\
        readline-dev\
        sqlite-dev\
        tcl-dev\
        tk\
        tk-dev\
        util-linux-dev\
        xz-dev\
        zlib-dev\
&& apk del --no-network .fetch-deps\
&& cd /usr/src/python\
&& gnuArch="$(dpkg-architecture --query DEB_BUILD_GNU_TYPE)"\
&& ./configure\
        --build="$gnuArch"\
        --enable-loadable-sqlite-extensions\
        --enable-optimizations\
        --enable-option-checking=fatal\
        --enable-shared\
        --with-system-expat\
        --with-system-ffi\
        --without-ensurepip\
&& make -j "$(nproc)"\
        EXTRA_CFLAGS="-DTHREAD_STACK_SIZE=0x100000"\
        LDFLAGS="-Wl,--strip-all"\
&& make install\
&& rm -rf /usr/src/python\
&& find /usr/local -depth\
        \(\
                \( -type d -a \( -name test -o -name tests -o -name idle_test \) \)\
                -o \( -type f -a \( -name '*.pyc' -o -name '*.pyo' -o -name '*.a' \) \)\
        \) -exec rm -rf '{}' +\
&& find /usr/local -type f -executable -not \( -name '*tkinter*' \) -exec scanelf --needed --nobanner --format '%n#p' '{}' ';'\
        | tr ',' '\n'\
        | sort -u\
        | awk 'system("[ -e /usr/local/lib/" $1 " ]") == 0 { next } { print "so:" $1 }'\
        | xargs -rt apk add --no-cache --virtual .python-rundeps\
&& apk del --no-network .build-deps\
&& python3 --version

RUN cd /usr/local/bin\
&& ln -s idle3 idle\
&& ln -s pydoc3 pydoc\
&& ln -s python3 python\
&& ln -s python3-config python-config

ENV PYTHON_PIP_VERSION=21.2.4

ENV PYTHON_SETUPTOOLS_VERSION=57.5.0

ENV PYTHON_GET_PIP_URL=https://github.com/pypa/get-pip/raw/d781367b97acf0ece7e9e304bf281e99b618bf10/public/get-pip.py

ENV PYTHON_GET_PIP_SHA256=01249aa3e58ffb3e1686b7141b4e9aac4d398ef4ac3012ed9dff8dd9f685ffe0

RUN set -ex;\
        wget -O get-pip.py "$PYTHON_GET_PIP_URL";\
echo "$PYTHON_GET_PIP_SHA256 *get-pip.py" | sha256sum -c -;\
        python get-pip.py\
        --disable-pip-version-check\
        --no-cache-dir\
        "pip==$PYTHON_PIP_VERSION"\
        "setuptools==$PYTHON_SETUPTOOLS_VERSION"\
;\
pip --version;\
        find /usr/local -depth\
        \(\
                \( -type d -a \( -name test -o -name tests -o -name idle_test \) \)\
                -o\
                \( -type f -a \( -name '*.pyc' -o -name '*.pyo' \) \)\
        \) -exec rm -rf '{}' +;\
rm -f get-pip.py

CMD ["python3"]
```
