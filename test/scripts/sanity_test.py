#!/usr/bin/python3

#from fabric.api import env, run
from fabric import Connection
import os
import vagrant


csiTestVM = "fedora29-csi-test-0.2"


def cleanVM(root):
    
    v = vagrant.Vagrant(root=root)
    print(" - Cleanig VM ", root) 

    try:
        v.destroy()
    except Exception as err:
        print(err)

    try:
        os.remove(root + "/Vagrantfile")
    except FileNotFoundError:
        pass

    

def initVM(vmName, root):
    buildPath = root + "/build"
    v = vagrant.Vagrant(root=root)

    if not os.path.exists(root):
        os.makedirs(root)

    print(" - Setting up VM ", root) 
    if not os.path.exists(buildPath):
        os.makedirs(buildPath)
    v.init(box_name=vmName)

def copyBins(bins, root):

    cmd = "cp -R {0}/* {1}/build/".format(bins,root)
    print(" - Copying binaries: ", cmd)
    os.system(cmd)


def runVM(root):
    v = vagrant.Vagrant(root=root)
    print(" - Starting VM ", root) 
    v.up()

def runPlugin(root):
    v = vagrant.Vagrant(root=root)

    # Start plugin
    cmd = "nohup /home/vagrant/build/jdss-csi-plugin --csi-address=127.0.0.1:15947 --soc-type=tcp --config ./build/controller-cfg.yaml >& /dev/null < /dev/null &"
    con = Connection(v.user_hostname_port(),
        connect_kwargs={
        "key_filename": v.keyfile(),
        })
    out = con.sudo(cmd)
    

def runCSISanity(root):
    v = vagrant.Vagrant(root=root)
    
    # Run tests
    print("Starting sanity tests.")
    #out = v.ssh(command="/home/vagrant/go/src/csi-test/cmd/csi-sanity/csi-sanity -ginkgo.failFast -csi.endpoint 127.0.0.1:15947")
    cmd = "/home/vagrant/go/src/csi-test/cmd/csi-sanity/csi-sanity -ginkgo.failFast -csi.endpoint 127.0.0.1:15947"
    print("Running: ", cmd)
    con = Connection(v.user_hostname_port(),
        connect_kwargs={
        "key_filename": v.keyfile(),
        })
    
    out = con.run(cmd)
    

def main():
    root = "csi-sanity"
    cleanVM(root)
    initVM(csiTestVM,root)
    copyBins("bins", root)
    try:
        runVM(root)
        runPlugin(root)
        runCSISanity(root)
    except Exception as err:
        print(err)
        raise err

    cleanVM(root)
    print("Success!")

if __name__ == "__main__":
    main()
