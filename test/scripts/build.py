#!/usr/bin/python3

import argparse
import os
import re
import shutil
import time
import vagrant
import subprocess
import sys

from fabric import Connection


vfile = """
# -*- mode: ruby -*-
# vi: set ft=ruby :
Vagrant.configure("2") do |config|
  config.vm.box = "<ubuntu-vm-name>"
  config.vm.synced_folder "../src", "/home/vagrant/go/src/github.com/open-e/JovianDSS-KubernetesCSI", owner: "vagrant", group:"vagrant"
  config.vm.synced_folder "../go", "/usr/local/go", owner: "vagrant", group:"vagrant"
  config.vm.provider "virtualbox" do |vb|
    config.vm.network "public_network", bridge: "enp1s0f1"
    vb.memory = "4096"
  end
  #config.vm.provision "shell" do |sh|
  #  sh.inline = "sudo sh -c 'echo GOROOT=/usr/local/go > /etc/profile.d/Z99-go.sh'"
  #  sh.inline = "sudo sh -c 'echo PATH=/usr/local/go/bin:$PATH >> /etc/profile.d/Z99-go.sh'"
  #  sh.inline = "mkdir -p /home/vagrant/go/src/github.com/open-e/JovianDSS-KubernetesCSI"
  #end
  config.vm.provision "ansible" do |ansible|
    ansible.playbook = "../src/test/env/build/ubuntu.yml"
  end
end
"""

def clean_vm(root):
    """Remove vagrant VM from specified path"""
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



def init_vm(name, root):
    """Init vagrant VM in given path"""

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
    v.halt()
    v.up()

def main(args):
    """Runs aggregation test on freshly build
            container of kubernetes csi plugin

    Parameters:
    root -- folder to run test in
    csi_test_vm -- name of vagrant VM to run test in
    """

    root = "./build"
    csi_test_vm = args.bvm

    clean_vm(root)

    init_vm(csi_test_vm, root)
    run_vm(root)

if __name__ == "__main__":

    parser = argparse.ArgumentParser()
    parser.add_argument('--no-clean', dest='nc', action='store_true',
            help='Do Not clean environment after execution.')
    parser.add_argument('--build-vm', dest='bvm', type=str, default="ubuntu/bionic64",
            help='VM template to be used for building plugin.')
    parser.add_argument('--branch', dest='branch', type=str, default="master",
            help='VM template to be used for building plugin.')

    args = parser.parse_args()
    main(args)
