# Deploying Fluxy to Kubernetes

You will need to build or load the weaveworks/fluxy image into the Docker daemon,
 since the deployment does not attempt to pull the image from a registry.
If you're using [minikube](https://github.com/kubernetes/minikube) to try things locally,
 for example, you can do

```
eval $(minikube docker-env)
make clean all
```

which will build the image in minikube's Docker daemon, thus making it
available to Kubernetes.

## Creating a key for automation

Fluxy updates a Git repository containing your Kubernetes config each
time a service is released; this will usually require an SSH access
key.

Here is an example of setting this up for the `helloworld` example in
the fluxy repository.

Fork the fluxy repository on github (you may also wish to rename it,
e.g., to `fluxy-testdata`). Now, we're going to add a deploy key so
fluxy can push to that repo. Generate a key in the console:

```
ssh-keygen -t rsa -b 4096 -f id-rsa-fluxy
```

This makes a private key file (`id-rsa-fluxy`) which we'll supply to
Fluxy in a minute, and a public key file (`id-rsa-fluxy.pub`) which
we'll now give to Github.

On the Github page for your forked repo, go to the settings and find
the "Deploy keys" page. Add one, check the write access box, and paste
in the contents of the `id-rsa-fluxy.pub` file -- the public key.

## Customising the deployment config

The file `fluxy-deployment.yaml` contains a Kubernetes deployment
configuration that runs the latest image of Fluxy.

You can create the deployment now:

```
kubectl create -f fluxy-deployment.yaml
```

To make the pod accessible to `fluxctl`, you can create a service for
Fluxy and use the Kubernetes API proxy to access it:

```
kubectl create -f fluxy-service.yaml
kubectl proxy &
export FLUX_URL=http://localhost:8001/api/v1/proxy/namespaces/default/services/fluxy
```

This will work with the default settings of `fluxctl`, and is
especially handy with minikube.

At this point you can see if it's all running by doing:

```
fluxctl list-services
```

To force Kubernetes to run the latest image after a rebuild, kill the pod:

```
kubectl get pods | grep fluxy | awk '{ print $1 }' | xargs kubectl delete pod
```

## Uploading a configuration

To begin using Fluxy, you need to provide at least the git repository
and the key from earlier.

Get a blank config with

```sh
fluxctl get-config > fluxy.conf
```

Now edit the file `fluxy.conf` -- it'll look like this:

```yaml
git:
  URL: ""
  path: ""
  branch: ""
  key: ""
slack:
  hookURL: ""
  username: ""
registry:
  auths: {}
```

Here's an example with values filled in:

```yaml
git:
  URL: git@github.com:squaremo/fluxy-testdata
  path: testdata
  branch: master
  key: |
         -----BEGIN RSA PRIVATE KEY-----
         ZNsnTooXXGagxg5a3vqsGPgoHH1KvqE5my+v7uYhRxbHi5uaTNEWnD46ci06PyBz
         zSS6I+zgkdsQk7Pj2DNNzBS6n08gl8OJX073JgKPqlfqDSxmZ37XWdGMlkeIuS21
         nwli0jsXVMKO7LYl+b5a0N5ia9cqUDEut1eeKN+hwDbZeYdT/oGBsNFgBRTvgQhK
         ... contents of id-rsa-fluxy file from above ...
         -----END RSA PRIVATE KEY-----
slack:
  hookURL: ""
  username: ""
registry:
  auths: {}
```

Note the use of `|` to have a multiline string value for the key; all
the lines must be indented if you use that.

If you use any private Docker image repositories, you will also need
to supply authentication information for those. These are given per
registry, like the following snippet (you can nick these from the
analagous section of ~/.docker/config.json, they are just
base64-encoded `<username>:<password>`):

```yaml
# ...
registry:
  auths:
    'https://index.docker.io/v1/':
      auth: "dXNlcm5hbWU6cGFzc3dvcmQK"
```

(NB the key is a URL, and will usually have to be quoted as it is above.)

Finally, give the config to Fluxy:

```sh
fluxctl set-config --file=fluxy.conf
```
