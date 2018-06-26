function analyze_solidity(){
    document.getElementById("analyze_result").innerHTML='<div class="hope">  <div class="am-g am-container">    <div class="am-u-lg-6 am-u-md-6 am-u-sm-12">        <p>Analyzing</p>    </div>  </div></div>';
    var type="solidity";
    var code=document.getElementById("solidity").value;
    var input=document.getElementById("solidity_input").value;
    httpPost(type, code, input);
}

function analyze_bytecode(){
    document.getElementById("analyze_result").innerHTML='<div class="hope">  <div class="am-g am-container">    <div class="am-u-lg-6 am-u-md-6 am-u-sm-12">        <p>Analyzing</p>    </div>  </div></div>';
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
        if (xmlhttp.readyState==4)
        {
            // 4 = "loaded"
            if (xmlhttp.status==200)
            {
                // 200 = OK
                // alert(xmlhttp.responseText);
                var result=xmlhttp.responseText;
                var result_show='<div class="hope">  <div class="am-g am-container">    <div class="am-u-lg-6 am-u-md-6 am-u-sm-12">        <p>';
                result_show+=JSON.parse(result)["data"].replace(/[\n\r]/g,'<br>');
                result_show+='</p>    </div>  </div></div>';
                document.getElementById("analyze_result").innerHTML=result_show;
            }
            else
            {
                alert("Problem retrieving XML data");
            }
        }

    }
}


