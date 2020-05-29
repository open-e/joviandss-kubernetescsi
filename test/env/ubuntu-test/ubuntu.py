#!/usr/bin/python3

import argparse
import time
import vagrant
import os
import re
import shutil
import subprocess
import sys

from fabric import Connection

test_name = "ubuntu-test"

vfile = """
# -*- mode: ruby -*-
# vi: set ft=ruby :
Vagrant.configure("2") do |config|
  config.vm.box = "<ubuntu-vm-name>"
  config.vm.synced_folder "../configs", "/configs"
  config.vm.synced_folder "../src", "/src"
  config.vm.provider "virtualbox" do |vb|
    config.vm.network "public_network", bridge: "enp1s0f1"
    vb.memory = "4096"
    vb.cpus = "2"
  end
  config.vm.provision "ansible" do |ansible|
    ansible.playbook = "main.yml"
  end
end
"""

def clean_vm(root):
    """Remove vagrant VM from specified path"""

    if os.path.isdir(root) == False:
        return

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

    clean_directory = ["rm", "-Rf", root]

    subprocess.check_output(clean_directory)

def init_vm(name, root):
    """Init vagrant VM in given path"""

    shutil.copytree("./src/test/env/" + test_name + "/", root)

    if not os.path.exists(root):
        os.makedirs(root)

    v = vagrant.Vagrant(root=root)

    print(" - Setting up VM ", root)
    v = vfile.replace("<ubuntu-vm-name>", name)
    f = open(root + "/Vagrantfile", 'w')
    f.write(v)
    f.close()

def run_vm(root):
    """Start vagrant VM"""

    v = vagrant.Vagrant(root=root)
    print(" - Starting VM ", root)
    v.up()

def get_version(src):
    """Get version of currently builded code """
    get_tag = ["git", "-C", src, "describe", "--long", "--tags"]
    tag_out = subprocess.check_output(get_tag)
    return tag_out.strip().decode('ascii')

def publish_container(root, argsi, version):
    """Publish tested container to dockerhub"""

    print(" - Publish container to dockerhub.")

    v = vagrant.Vagrant(root=root)

    con = Connection(v.user_hostname_port(),
                     connect_kwargs={
                         "key_filename": v.keyfile(),
                     })

    login_to_docker = ("docker login -u opene -p " + args.password)
    con.sudo(login_to_docker)

    if args.dpl == True:
        print(" - Publishing with tag latest.")
        set_tag_latest = ("docker tag opene/joviandss-csi-u:" + version +
                            " opene/joviandss-csi-u:latest")
        con.sudo(set_tag_latest)

        upload_latest = "docker push opene/joviandss-csi-u:latest"
        con.sudo(upload_latest)

    if args.dpv == True:
        print(" - Publishing with tag " + version)
        upload_latest = "docker push opene/joviandss-csi-u:" + version
        con.sudo(upload_latest)

    return

def main(args):
    """Runs aggregation test on freshly build
            container of kubernetes csi plugin

    Parameters:
    root -- folder to run test in
    csi_test_vm -- name of vagrant VM to run test in
    """
    csi_test_vm = args.tvm

    clean_vm(test_name)

    init_vm(csi_test_vm, test_name)

    run_vm(test_name)

    if (args.dpl or args.dpv):
        version = get_version("./src")
        publish_container(test_name, args, version)

    if args.nc == False:
        clean_vm(test_name)

    print("Success!")

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('--no-clean', dest='nc', action='store_true',
            help='Do Not clean environment after execution.')
    parser.add_argument('--test-vm', dest='tvm', type=str, default="ubuntu/bionic64",
            help='VM template to be used for building plugin.')
    parser.add_argument('--docker-pass', dest='password', type=str, default=None,
            help='Password for dockerhub.')
    parser.add_argument('--docker-push-latest', dest='dpl', action='store_true',
            help='Push container to tegistry as latest if tests are successful.')
    parser.add_argument('--docker-push-version', dest='dpv', action='store_true',
            help='Push container to tegistry according to src tag if tests are successful.')

    args = parser.parse_args()

    if (args.dpl or args.dpv) and (args.password == None):
        raise argparse.ArgumentTypeError('Please provide docker password for publishing.')

    main(args)
