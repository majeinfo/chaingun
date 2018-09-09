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
            data: [168.0]
        },
        
        {
            name: '#Errors',
            data: [0.0]
        },
        
        {
            name: '#Rcv Bytes',
            data: [42.0]
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
                  text: 'time(ms)'
               },
        },          
        
        series: [
        
        {
            name: 'Page upload',
            data: [168.0]
        },
        
        ]
            });
    var errors_by_code = Highcharts.chart('errors_by_code', {

	    title: {
            text: 'Error Codes per Second'
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
                  text: '#err'
               },
        },          
        
        series: [
        
        {
            name: '200',
            data: [1.0]
        },
        
        ]
            });
$('#http_codes > thead').append('<tr><th></th><th>200</th><th>#Req</th></tr>');
$('#http_codes > tbody:last-child').append('<tr><td>Page upload</td><td>1</td><td>1</td></tr>');
});
