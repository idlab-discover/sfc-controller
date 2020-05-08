# SFC-controller

The proposed SFC controller has been implemented as an extension to the default scheduling feature available in Kubernetes.
The SFC controller enables Kubernetes to efficiently allocate container-based service chains while maintaining bandwidth and
reducing latency.

## How does the SFC-controller work?
* In the `deployments` folder you find a Kubernetes Deployment that launches our SFC-controller. Information can be found in the [Kubernetes documentation](https://kubernetes.io/docs/tasks/administer-cluster/configure-multiple-schedulers/).
* Please see `main/main.go`. You should change the infrastructure weights/labels based on your own infrastructure before building the container.
* [Node labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) are used to make allocation decisions based on bandwidth and latency.
* Besides, our SFC-controller applies the new Kubernetes concepts of [pod and node informers](https://medium.com/@muhammet.arslan/write-your-own-kubernetes-controller-with-informers-9920e8ab6f84)
* If you would like to know further details about our SFC-controller, please read our papers mentioned below. 
 
## Build the SFC controller
1. Use the `Makefile` to build your own Docker image. `make build`.
2. Push the Docker image to a container registry: `make push`.

## Deploy the SFC controller
1. Deploy the SFC-controller policy configuration: `kubectl create -f deployments/scheduler-policy-config.yaml`.
4. Deploy the SFC-controller: `kubectl create -f deployments/sfc-controller-v2.yaml`.
5. Deploy your own pod. See `deployments/pod-example.yaml`.

## Documentation

PowerPoint presentations of this work are located [here](/docs). 

## Citation

If you use our work, please cite our articles.

* [NETSOFT 2019](https://ieeexplore.ieee.org/abstract/document/8806671?casa_token=6wRBKx50acMAAAAA:PXBO3-OBe1T9cXOIQpDge_L_vtxSM8pc7wHzXhkmOHAPPOnyTZ8FKmBzORRXEOx1BbU5dBp_)

```
@inproceedings{santos2019towards,
  title={Towards Network-Aware Resource Provisioning in Kubernetes for Fog Computing applications},
  author={Santos, Jos{\'e} and Wauters, Tim and Volckaert, Bruno and De Turck, Filip},
  booktitle={2019 IEEE Conference on Network Softwarization (NetSoft)},
  pages={351--359},
  year={2019},
  organization={IEEE}
}
```

* [SENSORS 2019](https://www.mdpi.com/1424-8220/19/10/2238)

```
@article{santos2019resource,
  title={Resource provisioning in Fog computing: From theory to practice},
  author={Santos, Jos{\'e} and Wauters, Tim and Volckaert, Bruno and De Turck, Filip},
  journal={Sensors},
  volume={19},
  number={10},
  pages={2238},
  year={2019},
  publisher={Multidisciplinary Digital Publishing Institute}
}
```

* [NOMS 2020](https://biblio.ugent.be/publication/8659903)

```
@inproceedings{pereira2020towards,
  title={Towards delay-aware container-based Service Function Chaining in Fog Computing},
  author={Pereira dos Santos, Jos{\'e} Pedro and Wauters, Tim and Volckaert, Bruno and De Turck, Filip},
  booktitle={NOMS2020, the IEEE/IFIP Network Operations and Management Symposium},
  pages={1--9},
  year={2020}
}
```

* [NETSOFT 2020](https://netsoft2020.ieee-netsoft.org/)

```
Demo presentation accepted: Live Demonstration of Service Function Chaining allocation in Fog Computing
```

## Team

* [Jose Santos](https://www.researchgate.net/profile/Jose_Santos171)

* [Tim Wauters](https://www.researchgate.net/profile/Tim_Wauters)

* [Bruno Volckaert](https://www.researchgate.net/profile/Bruno_Volckaert)

* [Filip de Turck](https://www.researchgate.net/profile/Filip_De_Turck)

## Contact

If you want to contribute, please contact:

Lead developer: [Jose Santos](https://github.com/jpedro1992/)

For questions or support, please use GitHub's issue system.

## License

Copyright (c) 2020 Ghent University and IMEC vzw.

Address: IDLab, Ghent University, iGent Toren, Technologiepark-Zwijnaarde 126 B-9052 Gent, Belgium 

Email: info@imec.be.
