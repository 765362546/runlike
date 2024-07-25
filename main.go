package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "strings"
    "os"

    "github.com/docker/docker/api/types"
    "github.com/docker/docker/client"
)

type Inspector struct {
    client    *client.Client
    container types.ContainerJSON
}

func NewInspector(containerID string) (*Inspector, error) {
    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        return nil, err
    }

    container, err := cli.ContainerInspect(context.Background(), containerID)
    if err != nil {
        return nil, err
    }

    return &Inspector{
        client:    cli,
        container: container,
    }, nil
}

func (i *Inspector) GetRunlikeCommand() string {
    command := "docker run "

    // Add container name
    containerName := i.container.Name[1:] // Strip leading '/'
    command += "--name " + containerName + " "

    // Add hostname
    //if i.container.Config.Hostname != "" {
    //    command += "--hostname " + i.container.Config.Hostname + " "
    //}

    // Add MAC address
    if i.container.NetworkSettings.MacAddress != "" {
        command += "--mac-address " + i.container.NetworkSettings.MacAddress + " "
    }

    // Add links
    for _, link := range i.container.HostConfig.Links {
        command += "--link " + link + " "
    }

    // Add cpuset-cpus
    if i.container.HostConfig.CpusetCpus != "" {
        command += "--cpuset-cpus " + i.container.HostConfig.CpusetCpus + " "
    }

    // Add cpuset-mems
    if i.container.HostConfig.CpusetMems != "" {
        command += "--cpuset-mems " + i.container.HostConfig.CpusetMems + " "
    }

    // Add devices
    for _, device := range i.container.HostConfig.Devices {
        command += "--device " + device.PathOnHost + ":" + device.PathInContainer
        if device.CgroupPermissions != "" {
            command += ":" + device.CgroupPermissions
        }
        command += " "
    }

    // Add labels
    //for key, value := range i.container.Config.Labels {
    //    command += "--label " + key + "=" + value + " "
    //}

    // Add memory
    if i.container.HostConfig.Memory > 0 {
        command += "--memory " + fmt.Sprintf("%d", i.container.HostConfig.Memory) + " "
    }

    // Add memory reservation
    if i.container.HostConfig.MemoryReservation > 0 {
        command += "--memory-reservation " + fmt.Sprintf("%d", i.container.HostConfig.MemoryReservation) + " "
    }

    // Add privileged
    if i.container.HostConfig.Privileged {
        command += "--privileged "
    }

    // Add detach (default to true for running containers)
    command += "-d "

    // Add tty
    if i.container.Config.Tty {
        command += "-t "
    }

    // Add rm (AutoRemove indicates whether to remove the container once it exits)
    if i.container.HostConfig.AutoRemove {
        command += "--rm "
    }

    // Add user
    if i.container.Config.User != "" {
        command += "--user " + i.container.Config.User + " "
    }

    // Add environment variables
    for _, env := range i.container.Config.Env {
        command += "-e \"" + env + "\" "
    }

    // Add port bindings
    for port, bindings := range i.container.HostConfig.PortBindings {
        for _, binding := range bindings {
            command += "-p " + binding.HostIP + ":" + binding.HostPort + ":" + string(port) + " "
        }
    }

    // Add network settings
    if i.container.HostConfig.NetworkMode != "" {
        command += "--network " + string(i.container.HostConfig.NetworkMode) + " "
    }

    // Add volumes
    for _, mount := range i.container.Mounts {
        command += "-v " + mount.Source + ":" + mount.Destination + " "
    }

    // Add working directory
    if i.container.Config.WorkingDir != "" {
        command += "-w " + i.container.Config.WorkingDir + " "
    }

    // Add restart policy
    if i.container.HostConfig.RestartPolicy.Name != "" {
        command += "--restart " + string(i.container.HostConfig.RestartPolicy.Name) + " "
    }

    // Add DNS settings
    for _, dns := range i.container.HostConfig.DNS {
        command += "--dns " + dns + " "
    }

    // Add DNS search domains
    for _, dnsSearch := range i.container.HostConfig.DNSSearch {
        command += "--dns-search " + dnsSearch + " "
    }

    // Add extra hosts
    for _, extraHost := range i.container.HostConfig.ExtraHosts {
        command += "--add-host " + extraHost + " "
    }

    // Always add the image name
    command += i.container.Config.Image + " "

    // Add the command and arguments
    if len(i.container.Config.Cmd) > 0 {
        command += strings.Join(i.container.Config.Cmd, " ") + " "
    }

    return command
}

func main() {
    containerID := flag.String("c", "", "name or id of the container")
    flag.Parse()

    if *containerID == "" {
        flag.PrintDefaults()
        os.Exit(0)
    }

    inspector, err := NewInspector(*containerID)
    if err != nil {
        log.Fatalf("Failed to create inspector: %v", err)
    }

    fmt.Println(inspector.GetRunlikeCommand())
}

