function analyze_solidity(){
    document.getElementById("start_button_1").innerHTML="<strong>Analyzing</strong>";
    var type="solidity";
    var code=document.getElementById("solidity").value;
    var input=document.getElementById("solidity_input").value;
    httpPost(type, code, input);
}

function analyze_bytecode(){
    document.getElementById("start_button_2").innerHTML="<strong>Analyzing</strong>";
    var type="bytecode";
    var code=document.getElementById("bytecode").value;
    var input=document.getElementById("bytecode_input").value;
    httpPost(type, code, input);
}


function httpPost(type, code, input) {
    var xmlhttp;
    xmlhttp=null;
    if (window.XMLHttpRequest)
    {
        // code for all new browsers
        xmlhttp=new XMLHttpRequest();
    }
    else if (window.ActiveXObject)
    {
        // code for IE5 and IE6
        xmlhttp=new ActiveXObject("Microsoft.XMLHTTP");
    }
    if (xmlhttp!=null)
    {
        xmlhttp.onreadystatechange=state_Change;
        xmlhttp.open("post","/api/analyze",true);
        //var content = "type="+type+"&code="+code+"&input="+input;
        var formData = new FormData();
        formData.append("type", type);
        formData.append("code", code);
        formData.append("input", input);
        xmlhttp.send(formData);
    }
    else
    {
        alert("Your browser does not support XMLHTTP.");
    }

    function state_Change(){
        document.getElementById("start_button_1").innerHTML="<strong>Click HERE to Start Analyzing!</strong>";
        document.getElementById("start_button_2").innerHTML="<strong>Click HERE to Start Analyzing!</strong>";
        if (xmlhttp.readyState==4)
        {
            // 4 = "loaded"
            if (xmlhttp.status==200)
            {
                // 200 = OK
                // alert(xmlhttp.responseText);
                var result=xmlhttp.responseText;
                var api_data=JSON.parse(result)
                var result_show='<div class="hope" style="background: '+api_data["color"]+'"><div class="am-g am-container"><div class="am-u-sm-12">';
                result_show+='<div class="hope-title">[Overflow Analysis Report]<br>'+api_data["title"]+'</div>';
                result_show+='<p>Analysis Log:<br><br>'+api_data["data"].replace(/[\n\r]/g,'<br>')+'</p>';
                result_show+='</div>  </div></div>';
                document.getElementById("analyze_result").innerHTML=result_show;
            }
            else
            {
                alert("Problem retrieving XML data");
            }
        }

    }
}


