# pbdemo

A tool to visualize and demo [kpack](https://github.com/pivotal/kpack). 

![Sample](docs/assets/sample.png)

#### Prerequisites

- Access to a kubernetes cluster with kpack installed.
- Cluster-admin permissions for the kubernetes cluster with kpack.
- Accessible Docker V2 Registry.

## Get Started

1. Download the newest [release](https://github.com/matthewmcnew/pbdemo/releases)
2. Run `pbdemo serve` to get an visualization of the images inside of a kpack cluster.   

### Demos

1. Start the local server for the build service visualization web UI

    ```bash
    pbdemo serve
    ```
    
    >  This should start up a local Build Service visualization web server that you access in the browser. 

1. Populate Build Service with sample image configurations.

    The `pbdemo populate` command will relocate builder and run images to a configured registry to enable build service demos.
    In addition, the command will "seed" an specified number of sample build service image configurations. 
    
    Running `pbdemo populate` will look something like this:
    ```bash
    pbdemo populate --registry gcr.io/my-project-name --count 20
    ```
   
    - `registry`: The registry to install build service images & for build service to build new images into. You need local write access to this registry.
    
    - `count`: The number of initial build service image configurations to create.
    
    - (Optional) `cache-size`: The Cache Size for each image's build cache. Example: `--cache-size 100Mi` Default: '500Mi'
    
    >  Warning: The registry configured in pbdemo populate must be publicly readable by kpack. 
    
1. Navigate to the Web UI in your browser to see build service build all the images created in step #3. 

## Demo: Stack Update

1. Navigate to the build service web UI and mark the current stack (run image) as 'vulnerable'.   

    - Copy the truncated stack digest from from one of the existing images in the visualization.
    - Click on Setup in the top right corner.
    - Paste the stack (run image) Digest into the Modal.
    - Click Save. 
    - You should see the images with that run image highlighted in red.  
     
1. Push an updated stack (Run Image)
    
    The `pbdemo update-stack` will push an updated image to the registry build service is monitoring. 
    
    ```
    pbdemo update-stack
    ```   

1. Navigate to the Web UI in your browser to watch build service `rebase` all the images that used the previous stack (run image).

## Demo: Buildpack update

1. Navigate to the build service web UI and mark a buildpack id & version as 'vulnerable'.   

    - Copy the current backpack ID & Version for from one of the existing images in the visualization.
    - Click on Setup in the top right corner.
    - Paste the Buildpack ID & Version into the Modal.
    - Click Save. 
    - You should see the images that were built with that buildpack highlighted in red.  
     
1. Push an Updated Backpack 
    
    The `pbdemo update-buildpacks --buildpack <buildpack>` will create a new buildpack and add it to the kpack buildpack store. Kpack will rebuild "out-of-date" images with the new buildpack.
    
    ```
    pbdemo update-buildpacks --buildpack <buildpack.id>
    ```
   
   Note: 

1. Navigate to the Web UI in your browser to watch build service `rebuild` all the images that used the previous buildpack.


## Image logs

You can view the build logs of any image in any namespace `pbdemo <image-logs>`.  

```
pbdemo logs <image-name>
```

## Cleanup
   
1. Remove all images created by `pbdemo` with `cleanup`

    ```
    pbdemo cleanup
    ```  
   
   Note: this will reset your kpack builder,stack, and store resources to their previous state before using pbdemo. 
