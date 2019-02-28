var elapsed_time = 2;
var total_requests = 6;
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
            data: [1.0, 1.0]
        },
        
        {
            name: '#Req',
            data: [3.0, 3.0]
        },
        
        {
            name: 'Latency (in ms)',
            data: [1.0, 1.0]
        },
        
        {
            name: '#Errors',
            data: [0.0, 0.0]
        },
        
        {
            name: '#Rcv Bytes',
            data: [83.0, 85.0]
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
                  text: 'time(ms)'
               },
        },          
        
        series: [
        
        {
            name: 'Page 1',
            data: [3.0, 2.0]
        },
        
        {
            name: 'Page 3',
            data: [1.0, 1.0]
        },
        
        {
            name: 'Page 4',
            data: [1.0, 1.0]
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
            categories: [0, 1],
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
            data: [3.0, 3.0]
        },
        
        ]
            });
$('#http_codes > thead').append('<tr><th></th><th>200</th><th>#Req</th></tr>');
$('#http_codes > tbody:last-child').append('<tr><td>Page 1</td><td>2</td><td>2</td></tr>');
$('#http_codes > tbody:last-child').append('<tr><td>Page 3</td><td>2</td><td>2</td></tr>');
$('#http_codes > tbody:last-child').append('<tr><td>Page 4</td><td>2</td><td>2</td></tr>');
});
