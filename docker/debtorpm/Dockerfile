FROM centos:7

# tools for repackaging deb to rpm
RUN yum install -y epel-release
RUN yum install -y alien
RUN yum --enablerepo=epel-testing install rpmrebuild -y

RUN yum clean all
