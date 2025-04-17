import React, { useState } from "react";
import DropdownMenu from "./DropDownMenu";

const Expense = ({ users, groupId, onExpenseAdded, onCancel }) => {
  const [checkedUsers, setCheckedUsers] = useState([]);
  const [expenseType, setExpenseType] = useState("");
  const [payerId, setPayerId] = useState("");
  const [amount, setAmount] = useState("");
  const [description, setDescription] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  if (!users) return null;

  const handleSubmit = async () => {
    // Validate form
    if (!description || !amount || !payerId || !expenseType || checkedUsers.length === 0) {
      alert("Please fill all required fields and select at least one user");
      return;
    }

    setIsLoading(true);

    const bodyData = {
      amount: parseInt(amount),
      payer_id: payerId,
      description: description,
      expense_type: expenseType,
      user_shares: checkedUsers.map((userId) => ({
        user_id: userId,
        share_amount: Number(document.getElementById(`amount-${userId}`).value),
      })),
    };

    try {
      const response = await fetch(`http://localhost:4000/api/addExpense/${groupId}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(bodyData),
      });
      
      if (!response.ok) {
        throw new Error('Failed to add expense');
      }
      
      const data = await response.json();
      console.log(data);
      
      // Call the callback if provided
      if (onExpenseAdded) {
        onExpenseAdded();
      } else {
        alert("Expense added successfully!");
        
        // Reset form if not being closed
        setCheckedUsers([]);
        setAmount("");
        setDescription("");
      }
      
    } catch (error) {
      console.error("Error adding expense:", error);
      alert("Failed to add expense. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  const toggle = (userId) => {
    setCheckedUsers((prevCheckedUsers) => {
      if (prevCheckedUsers.includes(userId)) {
        return prevCheckedUsers.filter((id) => id !== userId);
      } else {
        return [...prevCheckedUsers, userId];
      }
    });
  };

  return (
    <div className="space-y-3">
      {/* Two columns for top form elements */}
      <div className="grid grid-cols-2 gap-3">
        {/* Description - Full width */}
        <div className="col-span-2">
          <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-1">
            Description
          </label>
          <input 
            type="text" 
            id="description" 
            placeholder="What was this expense for?" 
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none"
          />
        </div>
        
        {/* Amount - First column */}
        <div>
          <label htmlFor="expense-amount" className="block text-sm font-medium text-gray-700 mb-1">
            Total Amount
          </label>
          <input 
            type="number" 
            id="expense-amount" 
            placeholder="0.00" 
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none"
          />
        </div>
        
        {/* Who paid? - Second column */}
        <div>
          <label htmlFor="payer" className="block text-sm font-medium text-gray-700 mb-1">
            Who Paid?
          </label>
          <DropdownMenu 
            options={users} 
            dropdownId="payer" 
            onSelect={(value) => setPayerId(value)}
          />
        </div>
      </div>
      
      {/* Split Type */}
      <div>
        <label htmlFor="expenseType" className="block text-sm font-medium text-gray-700 mb-1">
          Split Type
        </label>
        <DropdownMenu 
          options={[
            { label: "Split Equally", value: "EQUAL" },
            { label: "Exact Amounts", value: "EXACT" },
            { label: "Percentages", value: "PERCENTAGE" }
          ]} 
          dropdownId="expenseType"
          onSelect={(value) => setExpenseType(value)}
        />
      </div>
      
      {/* Users to split with */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Split With <span className="text-xs text-gray-500">({users.length} members)</span>
        </label>
        <div className="border border-gray-200 rounded-md">
          <div className="max-h-32 overflow-y-auto p-2">
            {users.map((user) => (
              <div key={user.id} className="flex items-center justify-between p-2 border border-gray-100 rounded hover:bg-gray-50 mb-2 last:mb-0">
                <div className="flex items-center flex-grow">
                  <input 
                    type="checkbox" 
                    id={`check-${user.id}`}
                    checked={checkedUsers.includes(user.id)}
                    onChange={() => toggle(user.id)} 
                    className="h-4 w-4 text-indigo-600 rounded mr-3"
                  />
                  <div className="flex-grow">
                    <div className="text-sm font-medium">{user.name}</div>
                    <div className="text-xs text-gray-500">{user.email}</div>
                  </div>
                </div>
                
                <input 
                  type="number" 
                  id={`amount-${user.id}`}
                  placeholder="0.00" 
                  disabled={!checkedUsers.includes(user.id)}
                  className="w-24 px-2 py-1 text-sm border border-gray-300 rounded shadow-sm"
                />
              </div>
            ))}
          </div>
        </div>
      </div>
      
      {/* Action Buttons */}
      <div className="flex justify-end space-x-3 pt-2">
        {onCancel && (
          <button
            type="button"
            onClick={onCancel}
            className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
          >
            Cancel
          </button>
        )}
        <button 
          onClick={handleSubmit}
          disabled={isLoading}
          className={`px-4 py-2 bg-indigo-600 text-white font-medium rounded-md hover:bg-indigo-700 focus:outline-none ${onCancel ? '' : 'w-full'}`}
        >
          {isLoading ? (
            <>
              <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white inline" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Processing...
            </>
          ) : (
            'Add Expense'
          )}
        </button>
      </div>
    </div>
  );
};

export default Expense;