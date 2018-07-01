var elapsed_time = 2;
var total_requests = 6;
var playbook_name = "test1-1VU.yml";

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
            categories: [0, 1],
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
            data: [2.0, 2.0]
        },
        
        {
            name: '#Req',
            data: [4.0, 2.0]
        },
        
        {
            name: 'Latency (in ms)',
            data: [328.0, 640.0]
        },
        
        {
            name: '#Errors',
            data: [1.0, 1.0]
        },
        
        {
            name: '#Rcv Bytes',
            data: [56.0, 16166.0]
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
            categories: [0, 1],
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
            data: [329.0, 0.0]
        },
        
        {
            name: 'Page 2',
            data: [327.0, 0.0]
        },
        
        {
            name: 'Page SSL',
            data: [0.0, 640.0]
        },
        
        ]
            });
});
