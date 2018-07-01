""" Creates Graphs from CSV
"""
import argparse
import sys
import os
import numpy as np
import pandas as pd

def graph(name, title, xaxis, xtitle, ytitle, series):
    output.write("    var %s = Highcharts.chart('%s', {\n" % (name, name))
    output.write('''
	    title: {
            text: '%s'
        },
        ''' % title)
    output.write('''
        legend: {
               layout: 'horizontal',
               align: 'center',
               verticalAlign: 'bottom',
               borderWidth: 0
            }, 
    ''')        
    output.write('''
        xAxis: {
            categories: %s,
            title: {
                text: '%s'
            },
        },
        ''' % (xaxis, xtitle))
    output.write('''
        yAxis: {
               title: {
                  text: '%s'
               },
        },          
        ''' % ytitle)        
    output.write('''
        series: [
        ''')
    for serie in series:
        output.write('''
        {
            name: '%s',
            data: %s
        },
        ''' % (serie['name'], serie['data']))
    output.write('''
        ]
        ''')
    output.write("    });\n")
	

parser = argparse.ArgumentParser(description='Create LoadTest Graphs.')
parser.add_argument('--data', dest='datafile',
                    help='Name of CSV file to process')
parser.add_argument('--script-name', dest='scriptname',
                    help='Name of Playbook file')
parser.add_argument('--output-dir', dest='outputdir',
                    help='Output Directory where to create the graphs')                    

args = parser.parse_args()

# Data file must be given
if args.datafile is None:
    print("CSV Filename is missing\n", file=sys.stderr)
    parser.print_help()
    exit(1)

# Output dir must be given
if args.outputdir is None:
    print("Output Directory is missing\n", file=sys.stderr)
    parser.print_help()
    exit(1)
    
# Create Output dir if necessary    
if os.path.isdir(args.outputdir):
	if not os.access(args.outputdir, os.W_OK):
		print("Directory %s is not writeable" % (args.outputdir,), file=sys.stderr)
		exit(1)
else:
	try:
		os.makedirs(args.outputdir)
	except Exception as e:
		print("Cannot create Directory %s: %s" % (args.outputdir, e))
		exit(1)
    
# Copy templates and js in output directory
base = os.path.dirname(sys.argv[0])
if os.system("cp -r " + base + "/templates/* '" + args.outputdir + "'"):
    print("Error while copying templates in %s" % (args.outputdir))
    exit(4)
    
# Read Data file    
try:
    results = pd.read_csv(args.datafile)
    #results.info()
except Exception as e:
    print("Error while reading data file %s: %s" %s (args.datafile, e))
    exit(2)
    
# Create Result file
try:
	outputfile = args.outputdir + os.path.sep + "data.js"
	output = open(outputfile, "w")
except Exception as e:
	print("Cannot create the output file %s: %s" % (outputfile, e))
	exit(3)
	
# Normalize data
results.Timestamp = results.Timestamp // 1_000_000_000
results.Latency = results.Latency // 1_000_000

# Compute stats
groupby_time = results.groupby('Timestamp')
x = pd.Series([t for t, list_idx in groupby_time])
y = pd.Series(np.zeros(len(x)))
nb_req = pd.Series(np.zeros(len(x)))
mean_time = pd.Series(np.zeros(len(x)))
errors = pd.Series(np.zeros(len(x)))
rcv_bytes = pd.Series(np.zeros(len(x)))

title_req = {}
groupby_title = results.groupby('Title')
for title, group in groupby_title:
    title_req[title] = pd.Series(np.zeros(len(x)))

idx = 0
total_elapsed_time = len(groupby_time)
total_requests = 0

for t, list_idx in groupby_time:
    vals = results.iloc[groupby_time.indices[t]]
    nb_req[idx] = len(list_idx)
    total_requests += len(list_idx)
    mean_time[idx] = int(vals['Latency'].mean())
    errors[idx] = len(np.where(vals['Status'] >= 400))
    rcv_bytes[idx] = int(np.sum(vals['Size']))
    groupby2 = vals.groupby('Vid')
    y[idx] = len(groupby2.groups)
    groupby3 = vals.groupby('Title')
    for title, group in groupby3:
        title_req[title][idx] = np.mean(group['Latency'])
        
    idx += 1

x -= x[0]

output.write('var elapsed_time = %d;\n' % total_elapsed_time)
output.write('var total_requests = %d;\n' % total_requests)
output.write('var playbook_name = "%s";\n\n' % args.scriptname)

output.write("$(function () {\n")

'''
graph(name='vu_per_second', title='#VU per second', xaxis=list(x), 
			xtitle='Elapsed Time (seconds)', ytitle='#VU',
		    series=[{'name': '#VU', 'data': list(y)}])
'''		    
		    
graph(name='overall_stats', title='Overall Statistics per Second', xaxis=list(x), 
			xtitle='Elapsed Time (seconds)', ytitle='',
		    series=[
				{'name': '#VU', 'data': list(y)},
				{'name': '#Req', 'data': list(nb_req)},
				{'name': 'Latency (in ms)', 'data': list(mean_time)},
				{'name': '#Errors', 'data': list(errors)},
				{'name': '#Rcv Bytes', 'data': list(rcv_bytes)},
			])		    

graph(name='stats_per_req', title='Latency per Request (in ms)', xaxis=list(x), 
			xtitle='Elapsed Time (seconds)', ytitle='',
		    series=[
		        {'name': title, 'data': list(group) } for (title, group) in title_req.items()
			])

output.write("});\n")
output.close()	

# EOF
