import { useState, useRef } from "react";
import {
  PiArrowFatLineUpFill,
  PiCaretRightFill,
  PiPaintBrushBroadBold,
} from "react-icons/pi";

export function FileSimulator() {
  const [fileContent, setFileContent] = useState("");
  const [output, setOutput] = useState("");
  const fileInputRef = useRef(null);

  const handleFileUpload = (event) => {
    const file = event.target.files[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => setFileContent(e.target.result);
      reader.readAsText(file);
    }
  };

  const handleExecute = async () => {
    //console.log(fileContent);
    const commands = fileContent
      .split("\n")
      .filter((command) => command.trim() !== "");
    let salida = "";

    for (const command of commands) {
      try {
        const response = await fetch("http://localhost:8080/analyze", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify([command]),
        });

        if (!response.ok) {
          const errorText = await response.text();
          salida += `Error en el servidor: ${errorText}\n`;
          continue;
        }

        const data = await response.json();

        // Agrega los resultados del comando al output
        if (data.results && typeof data.results === "object") {
          salida += `${Object.values(data.results).join("\n")}\n`;
        }

        // Agrega los errores del comando al output
        if(data.errors && typeof data.errors === 'object'){
          salida += `${Object.values(data.errors).map(e => `Error - ${e}`).join("\n")}\n`;
        }

      } catch (error) {
        salida += `Error en el servidor: ${error.message}\n`;
        console.log(error);
      }
    }

    setOutput(salida);

  };

  const handleClear = () => {
    setFileContent("");
  };

  return (
    <div className="flex flex-col items-center p-4 space-y-4 h-screen mr-4 ml-4">
      <h1 className=" bg-gray-800 text-white text-2xl font-bold p-3">
        Simulador de sistema de archivos EXT2 y EXT3
      </h1>

      <div className="flex space-x-2">
        <input
          type="file"
          ref={fileInputRef}
          onChange={handleFileUpload}
          className="hidden"
        />
        <button
          onClick={() => fileInputRef.current.click()}
          className="px-6 py-3 bg-green-500 text-white rounded hover:bg-green-700"
        >
          <PiArrowFatLineUpFill className="text-2xl" />
        </button>
        <button
          onClick={handleExecute}
          className="px-6 py-3 bg-blue-500 text-white rounded hover:bg-blue-700"
        >
          <PiCaretRightFill className="text-2xl" />
        </button>
        <button
          onClick={handleClear}
          className="px-6 py-3 bg-red-500 text-white rounded hover:bg-red-700"
        >
          <PiPaintBrushBroadBold className="text-2xl" />
        </button>
      </div>

      <div className=" bg-gray-800 h-full w-full p-6 rounded-md ">
        <textarea
          className=" bg-slate-300 w-full p-2 border rounded h-full"
          value={fileContent}
          onChange={(e) => setFileContent(e.target.value)}
          placeholder="Contenido del archivo..."
        />
      </div>

      <div className=" bg-gray-800 h-full w-full p-6 rounded-md">
        <textarea
          className=" bg-slate-300 w-full p-2 border rounded h-full outline-none border-none"
          value={output}
          onChange={(e) => setOutput(output)}
          readOnly
          placeholder="Salida..."
        />
      </div>
    </div>
  );
}
