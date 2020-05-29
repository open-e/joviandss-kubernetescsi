#!/usr/bin/python3

import argparse
import time
import os
import re
import shutil
import subprocess
import sys
import yaml


def register_container_in_vm(root, version):
    """Export fresh container to kubernetes cluster"""

    print(" - Adding container to the registry.")
    v = vagrant.Vagrant(root=root)
    cmd = "docker load < ./build/src/_output/joviandss-csi-u:" + version
    con = Connection(v.user_hostname_port(),
                     connect_kwargs={
                         "key_filename": v.keyfile(),
                     })
    con.sudo(cmd)

def get_version(src):
    """Get version of currently builded code """
    get_tag = ["git", "-C", src, "describe", "--long", "--tags"]
    tag_out = subprocess.check_output(get_tag)
    return tag_out.strip().decode('ascii')


def divide_yaml(f):
    """divide_yaml divides yaml on --- separated parts

        Takes file stream as an input and returns
        complete yaml data chunk or None if EOF found"""
    yaml_document_start = re.compile(r'^---$') 
    data = """"""
    for l in f.readline():
        if len(l) == 0:
            break
        if yaml_document_start.search(l):
            if len(data) > 0:
                yield data
                data = """"""
        
        yield "abc"

def specify_plugin_version(version, home):
    ctrl = "/configs/joviandss-csi-controller.yaml"


    y_ctrl = yaml.load(open(home + ctrl))
    #y_ctrl = subprocess.check_output(['ls', '~/'], shell=True)
    print(y_ctrl) 

def start_plugin(version):
    """Start controller and node plugins"""

    print(" - Starting plugin.")

    ctrl = "/src/deploy/joviandss/joviandss-csi-controller-u.yaml"

    node = "/src/deploy/joviandss/joviandss-csi-node-u.yaml"

    sc = "/home/kub/configs/joviandss-csi-sc.yaml"

    lctrl = "/home/kub/joviandss-csi-controller.yaml"
    lnode = "/home/kub/joviandss-csi-node.yaml"


    if subprocess.call(["cp"] + [ctrl] + [lctrl]):
        raise Exception("Unable to replicate controller def")

    if subprocess.call(["cp"] + [node] + [lnode]):
        raise Exception("Unable to replicate node def")

    never_load_plugin = [ "sed", "-i", "s/imagePullPolicy: Always/imagePullPolicy:  Never/g"]

    if subprocess.call(never_load_plugin + [lctrl]):
        raise Exception("Unable to omit controller plugin download from internet")

    if subprocess.call(never_load_plugin + [lnode]):
        raise Exception("Unable to omit node plugin download from internet")

    specify_plugin_version = [ "sed", "-i", "s/opene\/joviandss-csi-u:latest/opene\/joviandss-csi-u:"+  version + "/g"]
    if subprocess.call(specify_plugin_version + [lctrl]):
        raise Exception("Unable to specify controller plugin version")

    if subprocess.call(specify_plugin_version + [lnode]):
        raise Exception("Unable to specify node plugin version")

    kub_apply = ["kubectl", "apply", "-f"]

    if subprocess.call(kub_apply + [lctrl]):
        raise Exception("Unable to load controller plugin")

    if subprocess.call(kub_apply + [lnode]):
        raise Exception("Unable to load node plugin")

    if subprocess.call(kub_apply + [sc]):
        raise Exception("Unable to load storage class for plugin")


def start_nginx():
    """Start nginx plugin with added persistent volume"""

    print(" - Starting test deployment.")

    create_pvc = ["kubectl", "apply", "-f", "/src/deploy/example/nginx-pvc.yaml"]

    if subprocess.call(create_pvc):
        raise Exception("Unable to create pvc")

    start_nginx_cmd = ["kubectl", "apply", "-f", "/src/deploy/example/nginx.yaml"]

    if subprocess.call(start_nginx_cmd):
        raise Exception("Unable to load nginx plugin")

def wait_for_plugin_started(sec):
    """Wait for controller and node to start
        by scanning list of kubernetes pods
    """

    print(" - Waiting for plugin to start.")

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
        get_pods = ["kubectl", "get", "pods"]
        out = subprocess.check_output(get_pods).decode('ascii').splitlines()
        if not out:
            continue

        ctrl_running = ""
        node_running = ""

        for line in out:
            ctrl_running = ctrl_running_pattern.search(line)
            if ctrl_running is None:
                continue
            break

        for line in out:
            node_running = node_running_pattern.search(line)
            if node_running is None:
                continue
            break

        if ctrl_running != None and node_running != None:
            return True

        ctrl_creating = ""
        node_creating = ""
        for line in out:
            ctrl_creating = ctrl_creating_pattern.search(line)
            if ctrl_creating is None:
                continue
            break

        for line in out:
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
            get_pods = ["kubectl", "get", "pods"]
            out = subprocess.check_output(get_pods).decode('ascii')
            print(out)
            get_events = ["kubectl", "get", "events"]
            out = subprocess.check_output(get_events).decode('ascii')
            print(out)
            raise Exception("Fail during plugin loading.")

    raise Exception("Unable to get plugins to start running in time.")

def wait_for_nginx_started(sec):
    """Startn nginx container with JovianDSS volume
            and wait till it successfully loaded.
    """

    nginx_pending = re.compile(r'^nginx.*Pending.*$')
    nginx_running = re.compile(r'^nginx.*Running.*$')
    nginx_creating = re.compile(r'^nginx.*ContainerCreating.*$')

    while sec > 0:
        time.sleep(1)
        sec = sec - 1
        get_pods = ["kubectl", "get", "pods"]
        out = subprocess.check_output(get_pods).decode('ascii').splitlines()

        if not out:
            continue

        for line in out:
            found = nginx_running.search(line)
            if found is None:
                continue
            return True

        pending = None
        for line in out:
            pending = nginx_pending.search(line)
            if pending is None:
                continue
            break

        creating = None
        for line in out:
            creating = nginx_creating.search(line)
            if creating is None:
                continue
            break

        if (creating is None) and (pending is None):
            get_pods = ["kubectl", "get", "pods"]
            out = subprocess.check_output(get_pods).decode('ascii')
            print(out)
            get_events = ["kubectl", "get", "events"]
            out = subprocess.check_output(get_events).decode('ascii')
            print(out)
            raise Exception("Fail during nginx loading.")

    raise Exception("Unable to get nginx to start running in time.")

def main():
    """Runs aggregation test on freshly build
            container of kubernetes csi plugin

    Parameters:
    root -- folder to run test in
    csi_test_vm -- name of vagrant VM to run test in
    """
    version = get_version("/src")

    # Run tests section
    try:
#        specify_plugin_version('111', '/home/kub')
        start_plugin(version)
        wait_for_plugin_started(220)
        start_nginx()
        wait_for_nginx_started(120)
    except Exception as err:
        print(err)
        raise err

    print("Success!")

if __name__ == "__main__":
    parser = argparse.ArgumentParser()

    main()
