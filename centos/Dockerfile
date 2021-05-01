FROM centos:7

# Basic build tools
RUN yum -y install centos-release-scl
RUN yum -y install devtoolset-7-gcc-c++

# golang tools
RUN yum -y install wget
RUN wget https://golang.org/dl/go1.16.3.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.16.3.linux-amd64.tar.gz
RUN echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# dev libs for robotgo and fyne    
RUN yum -y install \
    libXcursor-devel \
    libXrandr-devel \
    libXinerama-devel \
    mesa-libGL-devel \
    libXtst-devel \
    libpng-devel \
    libxkbcommon-x11-devel

ENTRYPOINT ["scl", "enable", "devtoolset-7"]
CMD ["bash"]
