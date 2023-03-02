/* * * * * * * * * * * * * * * * * * * * * * * * * ** * * * * * * * * * * *
 * Battery widget plugin for freeboard.
 * Author: Vikas Lamba, https://github.com/vkylamba
 * Licensed under the MIT (http://raphaeljs.com/license.html) license.    *
 * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *  * * *
*/
(function () {

    const chartWidget = function (settings) {

        this.rootElement = null;
        this.currentSettings = settings;
        this.chartOptions = null;

        let chartHeight = this.currentSettings.chartHeight;
        let chartWidth = this.currentSettings.chartWidth;
        let background = this.currentSettings.background
  
        this.render = function (element) {
            let chartRootDiv = `
            <div
                id="${this.currentSettings.id}"
                style="width: ${chartWidth}px;height:${chartHeight}px;background:${background};"
                align="center">
              Chart
            </div>`;
            this.rootElement = $(chartRootDiv);
            $(element).append(this.rootElement);
        }
  
        this.onSettingsChanged = function (newSettings) {
            currentSettings = newSettings;
        }
  
        this.onCalculatedValueChanged = function (settingName, newValue) {
            console.log('onCalculatedValueChanged for ', settingName, newValue, this.chartOptions);
    
            if (settingName == 'data') {
                if (typeof newValue == String) {
                    this.chartOptions = JSON.parse(newValue);
                } else {
                    this.chartOptions = newValue;
                }
            }

            if (this.chartOptions != null) {
                // Initialize the echarts instance based on the prepared dom
                var myChart = echarts.init(document.getElementById(this.currentSettings.id));
                // Display the chart using the configuration items and data just specified.
                myChart.setOption(this.chartOptions);
            }
        }
  
        this.onDispose = function () {}
  
        this.getHeight = function () {
            return Number(currentSettings.height);
        }
  
        this.onSettingsChanged(settings);
    };

    freeboard.loadWidgetPlugin({
      "type_name": "chartWidget",
      "display_name": "E-Chart",    
      "fill_size": true,
      "external_scripts": [
        "https://cdn.jsdelivr.net/npm/echarts@5.4.1/dist/echarts.min.js"
      ],
      "settings": [
        {
          "name": "id",
          "display_name": "id",
          "default_value": "e-chart-1",
          "description": "DOM element id of the battery (must be unique)"
        },
        {
            "name": "id",
            "display_name": "background",
            "default_value": "",
            "description": "Chart background color"
        },
        {
          "name": "data",
          "display_name": "Chart Data",
          "type": "calculated",
          "description": "The chart data json"
        },    
        {
          "name": "chartHeight",
          "display_name": "Chart Height (px)",
          "type": "number",
          "default_value": 600,
          "description": "Chart height in pixels"
        },
        {
          "name": "chartWidth",
          "display_name": "Chart Width (px)",
          "type": "number",
          "default_value": 400,
          "description": "Chart width in pixels"
        },      
        {
          "name": "height",
          "display_name": "Height Blocks",
          "type": "number",
          "default_value": 7,
          "description": "A height block is around 60 pixels"
        }
      ],
      newInstance: function (settings, newInstanceCallback) {
        newInstanceCallback(new chartWidget(settings));
      }
    });
  
  }());