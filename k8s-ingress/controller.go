/*
Copyright 2015 The Kubernetes Authors All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"log"
	"os"
	"os/exec"
	"reflect"
	"text/template"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util"
)

const (
	nginxConf = `
events {
  worker_connections 1024;
}
http {
{{range $ing := .Items}}
{{range $rule := $ing.Spec.Rules}}
  server {
    listen 80;
    server_name {{$rule.Host}};
    resolver 127.0.0.1;
{{ range $path := $rule.HTTP.Paths }}
    {{if eq $path.Path "" }}
    location / {
    {{else}}
    location {{$path.Path}} {
    {{end}}
      proxy_pass http://{{$path.Backend.ServiceName}}:{{$path.Backend.ServicePort}}/;
    }{{end}}
  }{{end}}{{end}}
}`
)

func shellOut(cmd string) {
	log.Println("start nginx ", " cmd ", cmd)
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()

	log.Println(" cmd ", cmd, string(out))

	if err != nil {
//		log.Fatalf("Failed to execute %v: %v, err: %v", cmd, string(out), err)
		log.Printf("Failed to execute %v: %v, err: %v \n", cmd, string(out), err)
	}
}

func main() {
	log.SetFlags(log.Flags()|log.Lshortfile)

	log.Println("========>> 1")
	var ingClient client.IngressInterface
	if kubeClient, err := client.NewInCluster(); err != nil {
		log.Fatalf("Failed to create client: %v.", err)
	} else {
		ingClient = kubeClient.Extensions().Ingress(api.NamespaceAll)
	}
	tmpl, _ := template.New("nginx").Parse(nginxConf)
	log.Println("========>> 2")
	rateLimiter := util.NewTokenBucketRateLimiter(0.1, 1)
	known := &extensions.IngressList{}
	log.Println("========>> 3")
	// Controller loop
	shellOut("nginx")
	log.Println("========>> 4")
	for {
		log.Println("========>> 5")
		rateLimiter.Accept()
		log.Println("========>> 6")
		ingresses, err := ingClient.List(labels.Everything(), fields.Everything())
		if err != nil || reflect.DeepEqual(ingresses.Items, known.Items) {
			continue
		}
		log.Println("========>> 7")

		known = ingresses
		if w, err := os.Create("/etc/nginx/nginx.conf"); err != nil {
			log.Println("========>> 8")
			log.Fatalf("Failed to open %v: %v", nginxConf, err)
		} else if err := tmpl.Execute(w, ingresses); err != nil {
			log.Println("========>> 9")
			log.Fatalf("Failed to write template %v", err)
		}
		log.Println("========>> 10")
		shellOut("nginx -s reload")
	}
}