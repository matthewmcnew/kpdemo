# kpdemo

A tool to visualize and demo [kpack](https://github.com/pivotal/kpack). 

![Sample](docs/assets/sample.png)

#### Prerequisites

- Access to a kubernetes cluster with [kpack installed](https://github.com/pivotal/kpack/releases).
- Cluster-admin permissions for the kubernetes cluster with kpack.
- Accessible Docker V2 Registry.

## Get Started

1. Download the newest [release](https://github.com/matthewmcnew/kpdemo/releases)
2. Run `kpdemo serve` to get a visualization of the images inside of a kpack cluster.   

### Demos

1. Start the local server for the kpack visualization web UI

    ```bash
    kpdemo serve
    ```
    
    >  This should start up a local kpack visualization web server that you access in the browser. 

1. Populate kpack with sample image configurations.

    The `kpdemo populate` command will relocate builder and run images to a configured registry to enable kpack demos.
    In addition, the command will "seed" a specified number of sample kpack image configurations. 
    
    Running `kpdemo populate` will look something like this:
    ```bash
    kpdemo populate --registry gcr.io/my-project-name --count 20
    ```
   
    - `registry`: The registry to install kpack images & for kpack to build new images into. You need local write access to this registry.
    
    - `count`: The number of initial kpack image configurations to create.
    
    - (Optional) `cache-size`: The Cache Size for each image's build cache. Example: `--cache-size 100Mi` Default: '500Mi'
    
    >  Warning: The registry configured in kpdemo populate must be publicly readable by kpack. 
    
1. Navigate to the Web UI in your browser to see kpack build all the images created in step #3. 

## Demo: Stack Update

1. Navigate to the kpack web UI and mark the current stack (run image) as 'vulnerable'.   

    - Copy the truncated stack digest from from one of the existing images in the visualization.
    - Click on Setup in the top right corner.
    - Paste the stack (run image) Digest into the Modal.
    - Click Save. 
    - You should see the images with that run image highlighted in red.  
     
1. Push an updated stack (Run Image)
    
    The `kpdemo update-stack` will push an updated image to the registry kpack is monitoring. 
    
    ```
    kpdemo update-stack
    ```   

1. Navigate to the Web UI in your browser to watch kpack `rebase` all the images that used the previous stack (run image).

## Demo: Buildpack update

1. Navigate to the kpack web UI and mark a buildpack id & version as 'vulnerable'.   

    - Copy the current backpack ID & Version for from one of the existing images in the visualization.
    - Click on Setup in the top right corner.
    - Paste the Buildpack ID & Version into the Modal.
    - Click Save. 
    - You should see the images that were built with that buildpack highlighted in red.  
     
1. Push an Updated Backpack 
    
    The `kpdemo update-buildpack --buildpack <buildpack>` will create a new buildpack and add it to the kpack buildpack store. Kpack will rebuild "out-of-date" images with the new buildpack.
    
    ```
    kpdemo update-buildpack --buildpack <buildpack.id>
    ```
   

1. Navigate to the Web UI in your browser to watch kpack `rebuild` all the images that used the previous buildpack.


## Image logs

You can view the build logs of any image in any namespace `kpdemo <image-logs>`.  

```
kpdemo logs <image-name>
```

## Cleanup
   
1. Remove all images created by `kpdemo` with `cleanup`

    ```
    kpdemo cleanup
    ```  
   
   Note: this will reset your kpack builder,stack, and store resources to their previous state before using kpdemo. 
