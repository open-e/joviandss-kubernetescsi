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
    build_path = root + "/build"
    v = vagrant.Vagrant(root=root)

    if not os.path.exists(root):
        os.makedirs(root)

    print(" - Setting up VM ", root)
    if not os.path.exists(build_path):
        os.makedirs(build_path)
    v.init(box_name=name)

def run_vm(root):
    """Start vagrant VM"""

    v = vagrant.Vagrant(root=root)
    print(" - Starting VM ", root)
    v.up()

def init_test_env(src, root):
    """Create necessary resources in folder associated with test"""

    build = root + "/build"
    if not os.path.exists(build):
        os.makedirs(build)

    try:
        shutil.rmtree(root + "/build/src")
    except FileNotFoundError:
        pass

    dst = root + "/build/src"
    shutil.copytree(src, dst)

def load_modules(root):
    """Insert necessary kernel modules"""

    print(" - Loading modules.")
    v = vagrant.Vagrant(root=root)
    cmd = "modprobe iscsi_tcp"
    con = Connection(v.user_hostname_port(),
                     connect_kwargs={
                         "key_filename": v.keyfile(),
                     })
    con.sudo(cmd)

def register_container_in_vm(root, version):
    """Export fresh container to kubernetes cluster"""

    print(" - Adding container to the registry.")
    v = vagrant.Vagrant(root=root)
    cmd = "docker load < ./build/src/_output/joviandss-csi:" + version
    con = Connection(v.user_hostname_port(),
                     connect_kwargs={
                         "key_filename": v.keyfile(),
                     })
    con.sudo(cmd)

def add_secrets(root):
    """Add plugin configs as secrets to kubernetes"""

    print(" - Adding configs as secrets.")
    v = vagrant.Vagrant(root=root)
    add_controller_cmd = ("kubectl create secret generic jdss-controller-cfg "
                          "--from-file=./build/controller-cfg.yaml")
    add_node_cmd = ("kubectl create secret generic jdss-node-cfg "
                    "--from-file=./build/node-cfg.yaml")
    con = Connection(v.user_hostname_port(),
                     connect_kwargs={
                         "key_filename": v.keyfile(),
                     })
    con.run(add_controller_cmd)
    con.run(add_node_cmd)

def get_version(src):
    """Get version of currently builded code """
    get_tag = ["git", "-C", src, "describe", "--long", "--tags"]
    tag_out = subprocess.check_output(get_tag)
    return tag_out.strip().decode('ascii')

def start_plugin(root, version):
    """Start controller and node plugins"""

    print(" - Starting plugin.")
    v = vagrant.Vagrant(root=root)

    ctrl = "./build/src/deploy/joviandss/joviandss-csi-controller.yaml"

    node = "./build/src/deploy/joviandss/joviandss-csi-node.yaml"

    replace = "sed -i 's/imagePullPolicy: Always/imagePullPolicy:  IfNotPresent/g' "

    specify_version = "sed -i 's/opene\/joviandss-csi:latest/opene\/joviandss-csi:"+  version + "/g' "

    kub_apply = "kubectl apply -f "

    con = Connection(v.user_hostname_port(),
                     connect_kwargs={
                         "key_filename": v.keyfile(),
                     })
    con.run(replace + ctrl)
    con.run(replace + node)

    con.run(specify_version + ctrl)
    con.run(specify_version + node)

    con.run(kub_apply + ctrl)
    con.run(kub_apply + node)

def start_nginx(root):
    """Start nginx plugin with added persistent volume"""
    print(" - Starting test deployment.")
    v = vagrant.Vagrant(root=root)

    con = Connection(v.user_hostname_port(),
                     connect_kwargs={
                         "key_filename": v.keyfile(),
                     })
    create_pvc = "kubectl apply -f ./build/src/deploy/example/nginx-pvc.yaml"
    start_nginx_cmd = "kubectl apply -f ./build/src/deploy/example/nginx.yaml"
    con.run(create_pvc)
    con.run(start_nginx_cmd)


def wait_for_plugin_started(root, sec):
    """Wait for controller and node to start
        by scanning list of kubernetes pods
    """

    print(" - Waiting for plugin to start.")
    v = vagrant.Vagrant(root=root)

    con = Connection(v.user_hostname_port(),
                     connect_kwargs={
                         "key_filename": v.keyfile(),
                     })

    ctrl_running_pattern = re.compile(
        r'^joviandss-csi-controller-0.*3\/3.*Running.*$')
    ctrl_creating_pattern = re.compile(
        r'^joviandss-csi-controller-0.*ContainerCreating.*$')
    node_running_pattern = re.compile(
        r'^joviandss-csi-node-.*2\/2.*Running.*$')
    node_creating_pattern = re.compile(
        r'^joviandss-csi-node-.*ContainerCreating.*$')

    time.sleep(30)

    while sec > 0:
        sec = sec - 1
        time.sleep(1)
        out = str(con.run("kubectl get pods", hide=True).stdout)

        if not out:
            continue

        ctrl_running = ""
        node_running = ""

        for line in out.splitlines():
            ctrl_running = ctrl_running_pattern.search(line)
            if ctrl_running is None:
                continue
            break

        for line in out.splitlines():
            node_running = node_running_pattern.search(line)
            if node_running is None:
                continue
            break

        if ctrl_running != None and node_running != None:
            return True

        ctrl_creating = ""
        node_creating = ""
        for line in out.splitlines():
            ctrl_creating = ctrl_creating_pattern.search(line)
            if ctrl_creating is None:
                continue
            break

        for line in out.splitlines():
            node_creating = node_creating_pattern.search(line)
            if node_creating is None:
                continue
            break

        identified_statuses = len([i for i in [ctrl_creating,
                                               ctrl_running, 
                                               node_creating,
                                               node_running] if i != None])
        if identified_statuses != 2:
            print(identified_statuses)
            print([ctrl_creating, ctrl_running, node_creating, node_running])
            out = con.run("kubectl get pods")
            out = con.run("kubectl get events")
            raise Exception("Fail during plugin loading.")

    raise Exception("Unable to get plugins to start running in time.")

def wait_for_nginx_started(root, sec):
    """Startn nginx container with JovianDSS volume
            and wait till it successfully loaded.
    """

    v = vagrant.Vagrant(root=root)

    con = Connection(v.user_hostname_port(),
                     connect_kwargs={
                         "key_filename": v.keyfile(),
                     })
    nginx_pending = re.compile(r'^nginx.*Pending.*$')
    nginx_running = re.compile(r'^nginx.*Running.*$')
    nginx_creating = re.compile(r'^nginx.*ContainerCreating.*$')

    while sec > 0:
        time.sleep(1)
        sec = sec - 1
        out = str(con.run("kubectl get pods", hide=True).stdout)

        if not out:
            continue

        for line in out.splitlines():
            found = nginx_running.search(line)
            if found is None:
                continue
            return True

        pending = None
        for line in out.splitlines():
            pending = nginx_pending.search(line)
            if pending is None:
                continue
            break

        creating = None
        for line in out.splitlines():
            creating = nginx_creating.search(line)
            if creating is None:
                continue
            break

        if (creating is None) and (pending is None):
            print(out)
            out = con.run("kubectl get events")
            raise Exception("Fail during nginx loading.")

    raise Exception("Unable to get nginx to start running in time.")

def create_storage_class(root):
    """Create storage class in test VM"""

    print(" - Creating Storage class.")

    v = vagrant.Vagrant(root=root)

    sc_file = "./build/src/deploy/joviandss/joviandss-csi-sc.yaml"

    add_sc_cmd = "kubectl apply -f " + sc_file

    con = Connection(v.user_hostname_port(),
                     connect_kwargs={
                         "key_filename": v.keyfile(),
                     })
    con.run(add_sc_cmd)

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
        set_tag_latest = ("docker tag opene/joviandss-csi:" + version +
                            " opene/joviandss-csi:latest")
        con.sudo(set_tag_latest)

        upload_latest = "docker push opene/joviandss-csi:latest"
        con.sudo(upload_latest)

    if args.dpv == True:
        print(" - Publishing with tag " + version)
        upload_latest = "docker push opene/joviandss-csi:" + version
        con.sudo(upload_latest)

    return

def main(args):
    """Runs aggregation test on freshly build
            container of kubernetes csi plugin

    Parameters:
    root -- folder to run test in
    csi_test_vm -- name of vagrant VM to run test in
    """
    root = "aggregation-test"
    csi_test_vm = args.tvm
    clean_vm(root)

    init_vm(csi_test_vm, root)
    init_test_env("./build/src", root)
    version = get_version(root + "/build/src")

    # Run tests section
    try:
        run_vm(root)
        load_modules(root)
        register_container_in_vm(root, version)
        add_secrets(root)
        start_plugin(root, version)
        wait_for_plugin_started(root, 220)
        create_storage_class(root)
        start_nginx(root)
        wait_for_nginx_started(root, 120)
    except Exception as err:
        print(err)
        raise err

    # Publish section
    if (args.dpl or args.dpv):
        publish_container(root, args, version)

    if args.nc == True:
        clean_vm(root)

    print("Success!")

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('--no-clean', dest='nc', action='store_true',
            help='Do Not clean environment after execution.')
    parser.add_argument('--test-vm', dest='tvm', type=str, default="kubernetes-14.3",
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
