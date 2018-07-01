import json
import os.path
import glob
from django.shortcuts import render
from django.http import HttpResponse
from django.views.decorators.csrf import csrf_exempt
from .settings import REPOSITORY_ROOTDIR

# Memorize the script name played by injectors
script_names = {}
script_name = ''


def index(request):
    context = {}
    return render(request, 'index.html', context)


@csrf_exempt
def clean_results(request):
    script_names = {}
    script_name = ''
    return HttpResponse(json.dumps({ 'status': 'OK', 'msg': '' }), content_type='application/json')


@csrf_exempt
def store_results(request, repository, resultname, scriptname):
    global script_names, script_name
    
    context = {}

    # Create repository if needed
    repo_dir = os.path.join(REPOSITORY_ROOTDIR, repository)
    if not os.path.isdir(repo_dir):
        try:
            os.mkdir(repo_dir)
        except Exception as e:
            return HttpResponse(json.dumps({ 'status': 'Error', 'msg': str(e) }), content_type='application/json')

    fname = os.path.join(repo_dir, resultname)
    try:
        with open(fname, 'wb') as f:
            f.write(request.body)
    except Exception as e:
        return HttpResponse(json.dumps({ 'status': 'Error', 'msg': str(e) }), content_type='application/json')

    script_names[resultname] = scriptname
    script_name = scriptname

    return HttpResponse(json.dumps({ 'status': 'OK', 'msg': '' }), content_type='application/json')


@csrf_exempt
def merge_results(request, repository):
    repo_dir = os.path.join(REPOSITORY_ROOTDIR, repository)
    merged_name = 'merged.csv'
    fname = os.path.join(repo_dir, merged_name)
    try:
        with open(fname, 'wb') as f:
            first_file = True
            for res in glob.glob(os.path.join(repo_dir, '*')):
                # Avoid recursivity !
                if os.path.basename(res) == merged_name:
                    continue
                first_line = True
                try:
                    with open(res, 'rb') as r:
                        for line in r:
                            if first_line and not first_file:
                                first_line = False
                                continue
                            f.write(line)
                    first_file = False
                except Exception as e:
                    return HttpResponse(json.dumps({ 'status': 'Error', 'msg': str(e) }), content_type='application/json')
        
        # Call the external Python viewer.py like the standalone player does...
        # TODO: we display only one script name even if the injectors played different scripts...
        cmd = f"python ../../python/viewer.py --data '{repo_dir}/{merged_name}' --script-name '{script_name}' --output-dir '{repo_dir}'"
        #print(cmd)
        status = os.system(cmd)
        if status:
            return HttpResponse(json.dumps({ 'status': 'Error', 'msg': 'Graph generation failed' }), content_type='application/json')

    except Exception as e:
        return HttpResponse(json.dumps({ 'status': 'Error', 'msg': str(e) }), content_type='application/json')

    return HttpResponse(json.dumps({ 'status': 'OK', 'msg': '', 'link_url': f'/static/data/{repository}/index.html' }), content_type='application/json')


def _create_graphs(repo_dir):
    pass


