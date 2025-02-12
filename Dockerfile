FROM redos:7.3.4
#CMD sleep 1000000000000000
#ENV startcmd = "-1 -p:/home/dialog/mpslin1/stnlink/prj/dc2/r.stnlink"
ENV GIGACHAT_AUTH_DATA="MDU2MGFjMGItN2M4Yi00Mjc0LTk4NmEtNzA4ODhhODM0YWE5OjVhZjBhN2I3LTM2OTUtNDk3Ni04YjY0LWMwMzViOTc5NjQzMQ=="
ENV TELEGRAM_TOKEN="6295776071:AAEtsV2F3sqynbD2bDxJw1EMe8HrpgHsw34"
ENV TZ="Europe/Moscow"
ADD bin /bin
#RUN rpm -i /bin/lp1Apkan-1.1.2-1.el7.noarch.rpm /distro/stnlink-2.1.10-1.el7.x86_64.rpm 
#CMD /opt/dialog/stnlink/stnlink.sh -1 -p:/opt/dialog/pu1Apkan/r.stnlink
CMD  ln -s /bin/userslist /userslist && ./bin/telegrambot
#CMD ./bin/telegrambot
#COPY mpslin1 /app
#RUN [A"/opt/dialog/stnlink/stnlink.sh", "-1", "-p:/opt/dialog/lp1Apkan/r.stnlink"]
