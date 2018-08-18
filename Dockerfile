FROM scratch

ENTRYPOINT [ "/ipvs-operator" ]

ADD ipvs-operator /ipvs-operator
