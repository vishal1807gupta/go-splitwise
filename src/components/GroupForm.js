import React from "react";

const GroupForm = ({ onSuccess, onClose }) => {
    const handleSubmit = async (e) => {
        e.preventDefault();
        const groupName = document.getElementById("groupName").value;
        
        try {
            const response = await fetch("http://localhost:4000/api/creategroup", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ group_name: groupName }),
            });
            
            const data = await response.json();
            console.log(data);
            
            // Instead of navigating, call the onSuccess prop if provided
            if (onSuccess) {
                onSuccess(data);
            }
            
            // Close the modal
            if (onClose) {
                onClose();
            }
        } catch (error) {
            console.error("Error creating group:", error);
        }
    };
    
    return (
        <div className="w-full">
            <div className="mb-4">
                <input 
                    type="text" 
                    placeholder="Group Name" 
                    id="groupName"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-pink-500 focus:border-pink-500"
                ></input>
            </div>
            <div className="flex justify-end">
                <button 
                    onClick={handleSubmit} 
                    className="bg-green-500 hover:bg-green-700 text-white font-bold py-2 px-4 rounded"
                >
                    Submit
                </button>
            </div>
        </div>
    );
};

export default GroupForm;