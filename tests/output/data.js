var elapsed_time = 1;
var total_requests = 1;
var playbook_name = "None";

$(function () {
    var overall_stats = Highcharts.chart('overall_stats', {

	    title: {
            text: 'Overall Statistics per Second'
        },
        
        legend: {
               layout: 'horizontal',
               align: 'center',
               verticalAlign: 'bottom',
               borderWidth: 0
            }, 
    
        xAxis: {
            categories: [0],
            title: {
                text: 'Elapsed Time (seconds)'
            },
        },
        
        yAxis: {
               title: {
                  text: ''
               },
        },          
        
        series: [
        
        {
            name: '#VU',
            data: [1.0]
        },
        
        {
            name: '#Req',
            data: [1.0]
        },
        
        {
            name: 'Latency (in ms)',
            data: [3.0]
        },
        
        {
            name: '#Errors',
            data: [1.0]
        },
        
        {
            name: '#Rcv Bytes',
            data: [14.0]
        },
        
        ]
            });
    var stats_per_req = Highcharts.chart('stats_per_req', {

	    title: {
            text: 'Latency per Request (in ms)'
        },
        
        legend: {
               layout: 'horizontal',
               align: 'center',
               verticalAlign: 'bottom',
               borderWidth: 0
            }, 
    
        xAxis: {
            categories: [0],
            title: {
                text: 'Elapsed Time (seconds)'
            },
        },
        
        yAxis: {
               title: {
                  text: ''
               },
        },          
        
        series: [
        
        {
            name: 'Page 1',
            data: [3.0]
        },
        
        ]
            });
});
