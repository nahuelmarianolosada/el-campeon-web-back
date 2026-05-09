/**
 * API Helpers para Checkout Frontend
 *
 * Este archivo proporciona funciones helper para integrar el nuevo sistema de checkout
 * con el backend mejorado que soporta múltiples métodos de pago y entrega.
 *
 * Ubicación recomendada: lib/api.ts o similar en tu proyecto Next.js
 */

export interface CreateOrderPayload {
  shipping_address: {
    street: string
    city: string
    postal_code: string
    country: string
  }
  delivery_method: "shipping" | "pickup-libreria" | "pickup-jugueteria"
  notes?: string
}

export interface CreatePaymentPayload {
  order_id: number
  amount: number
  payment_method: "MP_SAVED" | "MP_INSTALLMENTS" | "MP_CARD" | "CASH"
}

export interface OrderResponse {
  id: number
  order_number: string
  user_id: number
  status: string
  subtotal: number
  tax: number
  total: number
  shipping_address: Record<string, any>
  delivery_method: "shipping" | "pickup-libreria" | "pickup-jugueteria"
  items: Array<{
    id: number
    product_id: number
    quantity: number
    price: number
    subtotal: number
    product: {
      id: number
      name: string
      description?: string
      image_url?: string
      [key: string]: any
    }
  }>
  notes?: string
  created_at: string
  updated_at: string
}

export interface PaymentResponse {
  id: number
  transaction_id: string
  order_id: number
  user_id: number
  amount: number
  currency: string
  status: string
  payment_method: string
  mercadopago_preference_id?: string
  mercadopago_data?: Record<string, any>
  approved_at?: string
  rejected_reason?: string
  created_at: string
  updated_at: string
}

/**
 * Crear una nueva orden
 *
 * @param token - JWT token del usuario autenticado
 * @param payload - Datos de la orden (dirección, método de entrega, notas)
 * @returns - Respuesta con los datos de la orden creada
 *
 * @example
 * const order = await createOrder(token, {
 *   shipping_address: {
 *     street: "Calle Principal 123",
 *     city: "Buenos Aires",
 *     postal_code: "1425",
 *     country: "Argentina"
 *   },
 *   delivery_method: "shipping",
 *   notes: "Por favor dejar en portería"
 * })
 */
export async function createOrder(
  token: string,
  payload: CreateOrderPayload
): Promise<OrderResponse> {
  const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/orders`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(payload),
  })

  if (!response.ok) {
    const error = await response.json()
    throw new Error(error.error || "Error creating order")
  }

  return response.json()
}

/**
 * Crear un nuevo pago para una orden
 *
 * @param token - JWT token del usuario autenticado
 * @param payload - Datos del pago (orden, monto, método de pago)
 * @returns - Respuesta con los datos del pago creado
 *
 * @example
 * const payment = await createPayment(token, {
 *   order_id: order.id,
 *   amount: order.total,
 *   payment_method: "MP_CARD"
 * })
 *
 * // Si es MercadoPago, redirigir al usuario a la preference URL
 * if (payment.mercadopago_preference_id) {
 *   window.location.href = \`https://www.mercadopago.com.ar/checkout/v1/redirect?pref_id=\${payment.mercadopago_preference_id}\`
 * }
 */
export async function createPayment(
  token: string,
  payload: CreatePaymentPayload
): Promise<PaymentResponse> {
  const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/payments`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(payload),
  })

  if (!response.ok) {
    const error = await response.json()
    throw new Error(error.error || "Error creating payment")
  }

  return response.json()
}

/**
 * Obtener información de un pago por ID de orden
 *
 * @param token - JWT token del usuario autenticado
 * @param orderId - ID de la orden
 * @returns - Datos del pago asociado a la orden
 */
export async function getPaymentByOrderId(
  token: string,
  orderId: number
): Promise<PaymentResponse> {
  const response = await fetch(
    `${process.env.NEXT_PUBLIC_API_URL}/api/payments/order/${orderId}`,
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  )

  if (!response.ok) {
    const error = await response.json()
    throw new Error(error.error || "Error fetching payment")
  }

  return response.json()
}

/**
 * Obtener la orden creada (para verificar su estado)
 *
 * @param token - JWT token del usuario autenticado
 * @param orderId - ID de la orden
 * @returns - Datos de la orden
 */
export async function getOrder(token: string, orderId: number): Promise<OrderResponse> {
  const response = await fetch(
    `${process.env.NEXT_PUBLIC_API_URL}/api/orders/${orderId}`,
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  )

  if (!response.ok) {
    const error = await response.json()
    throw new Error(error.error || "Error fetching order")
  }

  return response.json()
}

/**
 * Helper: Obtener dirección depickup según el método de entrega
 *
 * @param deliveryMethod - Método de entrega seleccionado
 * @returns - Objeto con los datos de dirección
 */
export function getPickupAddress(
  deliveryMethod: "shipping" | "pickup-libreria" | "pickup-jugueteria"
) {
  const addresses = {
    "pickup-libreria": {
      street: "Güemes 901, San Salvador de Jujuy, Jujuy, Argentina",
      city: "San Salvador de Jujuy",
      postal_code: "4600",
      country: "Argentina",
    },
    "pickup-jugueteria": {
      street: "Güemes 1045, San Salvador de Jujuy, Jujuy, Argentina",
      city: "San Salvador de Jujuy",
      postal_code: "4600",
      country: "Argentina",
    },
  }

  return addresses[deliveryMethod as keyof typeof addresses]
}

/**
 * Helper: Procesar el pago después de crear la orden
 *
 * Maneja diferentes tipos de pago:
 * - MP_SAVED, MP_INSTALLMENTS, MP_CARD: Redirige a MercadoPago
 * - CASH: Muestra código de pago para Pago Fácil/Rapipago
 *
 * @param payment - Datos del pago creado
 * @param paymentMethod - Método de pago utilizado
 */
export async function processPayment(
  payment: PaymentResponse,
  paymentMethod: "MP_SAVED" | "MP_INSTALLMENTS" | "MP_CARD" | "CASH"
) {
  if (["MP_SAVED", "MP_INSTALLMENTS", "MP_CARD"].includes(paymentMethod)) {
    // Redirigir a MercadoPago
    if (payment.mercadopago_preference_id) {
      // URL del checkout web de MercadoPago
      const preferenceUrl = `https://www.mercadopago.com.ar/checkout/v1/redirect?pref_id=${payment.mercadopago_preference_id}`
      window.location.href = preferenceUrl
    } else {
      throw new Error("No MercadoPago preference ID received")
    }
  } else if (paymentMethod === "CASH") {
    // Para pagos en efectivo, mostrar código de pago
    return {
      type: "cash",
      paymentCode: payment.transaction_id,
      message: `Código de pago: ${payment.transaction_id}. Presentalo en Pago Fácil o Rapipago.`,
    }
  }
}

/**
 * EJEMPLO DE USO COMPLETO EN UN COMPONENTE REACT:
 *
 * ```tsx
 * "use client"
 * import { createOrder, createPayment, processPayment } from "@/lib/api"
 *
 * export default function CheckoutPage() {
 *   const handleCheckout = async (e: React.FormEvent) => {
 *     e.preventDefault()
 *     try {
 *       // 1. Crear la orden
 *       const order = await createOrder(token, {
 *         shipping_address: {},
 *         delivery_method: "shipping",
 *         notes: "..."
 *       })
 *
 *       // 2. Crear el pago
 *       const payment = await createPayment(token, {
 *         order_id: order.id,
 *         amount: order.total,
 *         payment_method: "MP_CARD"
 *       })
 *
 *       // 3. Procesar el pago (redirige a MercadoPago o muestra código)
 *       await processPayment(payment, "MP_CARD")
 *     } catch (error) {
 *       console.error(error)
 *     }
 *   }
 * }
 * ```
 */

